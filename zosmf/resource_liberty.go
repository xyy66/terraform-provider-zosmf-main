package zosmf

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// "strconv"

	"log"
	"net/http"

	"bytes"

	"io/ioutil"

	"encoding/json"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
)

type createWorkflowResponse struct {
	WorkflowKey         string `json:"workflowKey"`
	Vendor              string `json:"vendor"`
	WorkflowVersion     string `json:"workflowVersion"`
	WorkflowDescription string `json:"workflowDescription"`
	WorkflowID          string `json:"workflowID"`
}

type getWfPropertiesResponse struct {
	WorkflowName string `json:"workflowName"`
	WorkflowKey  string `json:"workflowKey"`
	StatusName   string `json:"statusName"`
}

type liberty struct {
}

func resourceLiberty() *schema.Resource {
	log.Printf("[DEBUG] Called func %s", "resourceLiberty")
	return &schema.Resource{
		CreateContext: resourceLibertyCreate,
		ReadContext:   resourceLibertyRead,
		UpdateContext: resourceLibertyUpdate,
		DeleteContext: resourceLibertyDelete,
		Schema: map[string]*schema.Schema{
			"workflow_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"wf_instance_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"create_liberty_workflow_path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"destroy_liberty_workflow_path": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"system": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"create_liberty_variable_file_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"destroy_liberty_variable_file_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"wf_variables": &schema.Schema{
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
			"running_liberty": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceLibertyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "resourceLibertyCreate")
	createLibertyWfDefPath := d.Get("create_liberty_workflow_path").(string)
	createLibertyWfVarPath := d.Get("create_liberty_variable_file_path").(string)
	workflow_instance_name := d.Get("wf_instance_name").(string) + " start"

	wfKey, diags := createAWorkflow(ctx, d, createLibertyWfDefPath, createLibertyWfVarPath, m, workflow_instance_name)
	if wfKey == "" {
		log.Printf("[SEVERE] Create a Workflow failed!")
		return diags
	}
	d.Set("workflow_id", wfKey)
	d.Set("Hostname", "myLocalHost")
	responseCode, diags := startAWorkflow(ctx, d, wfKey, m)
	if responseCode != 202 {
		log.Printf("[SEVERE] Start Workflow key: %s, failed，responsed code: %d!", wfKey, responseCode)
		return diags
	}
	fetchStatusCount := 0
	completeSuccessfully := false
	for {
		time.Sleep(5 * time.Second)
		wfStatus, _ := getPropertiesOfAWorkflow(ctx, d, wfKey, m)
		if wfStatus == "complete" {
			log.Printf("[DEBUG] Workflow workflowKey : %s is complete.", wfKey)
			completeSuccessfully = true
			break
		} else if wfStatus == "automation-in-progress" {
			log.Printf("[DEBUG] Workflow workflowKey : %s is in automation.", wfKey)
		} else {
			log.Printf("[DEBUG] The automation of workflow workflowKey : %s is stop, but the workflow is not complete.", wfKey)
			break
		}
		fetchStatusCount++
		if fetchStatusCount > 3 {
			break
		}
	}

	resourceID := uuid.New().String()
	hostName := defaultHost
	variableMap := d.Get("wf_variables").(map[string]interface{})
	if variableMap == nil {
		log.Printf("[DEBUG] Cannot get the workflow variable for host name, use the default value!")
	} else {
		hostName = variableMap["zcx_ipv4"].(string)
		if hostName == "" {
			log.Printf("[DEBUG] Cannot get the host name from workflow variable, use the default value!")
			hostName = defaultHost
		}
	}
	hostPort := defaultPort
	url := "https://" + hostName + ":" + hostPort
	runLiberty := map[string]string{"ID": resourceID, "Hostname": hostName, "Hostport": hostPort, "url": url}
	d.Set("running_liberty", runLiberty)

	if completeSuccessfully {
		d.SetId(resourceID)
		return diags
	} else {
		fmt.Println("The workflow is not complete")
		d.SetId(resourceID)
		return diags
	}
}
func createAWorkflow(ctx context.Context, d *schema.ResourceData, wfDefPath string, wfVariablePath string, m interface{}, wf_instance_name string) (string, diag.Diagnostics) {
	log.Printf("[DEBUG] Called func %s", "createAWorkflow")
	client := m.(Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(10000)
	instance_name := fmt.Sprintf("%s_%d", wf_instance_name, randomNum)

	log.Printf("[DEBUG] Instance name     ... %s  ", instance_name)

	zosmf_workflow_definition_path_full := wfDefPath
	system := d.Get("system")
	workflow_variable_file_path := wfVariablePath
	variableArrStr := getVariableArrayStr(d) // get the variable array string, format like: "variables":[{"name":"st_user","value":"ZOSMFT1"}]

	var jsonStr = []byte(fmt.Sprintf(`{
            "workflowName":"%s",
            "workflowDefinitionFile":"%s",
            "system":"%s",
            "owner":"%s",
            "assignToOwner":true,
            "variableInputFile":"%s",
			%s
		}`, instance_name, zosmf_workflow_definition_path_full, system, client.Username, workflow_variable_file_path, variableArrStr))

	zosmf_workflow_uri_full := fmt.Sprintf("%s/workflow/rest/1.0/workflows", client.HostURL)
	log.Printf("[DEBUG] zOSMF Workflow full path: %s", zosmf_workflow_uri_full)
	req, err := http.NewRequest(
		"POST",
		zosmf_workflow_uri_full,
		bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))
	req.AddCookie(&http.Cookie{Name: "LtpaToken2", Value: client.Token, Path: "/", HttpOnly: true})
	log.Printf("[DEBUG] zosmf create workflow request body: %s", req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] zosmf workflow creation response Body: %s", string(body))

	workflow := new(createWorkflowResponse)

	err = json.Unmarshal(body, &workflow)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[DEBUG] Workflow key: %s", workflow.WorkflowKey)
	d.Set("workflow_id", workflow.WorkflowKey)

	//d.SetId(uuid.New().String())

	var wfKey string = workflow.WorkflowKey
	return wfKey, diags
}

func getVariableArrayStr(d *schema.ResourceData) string {
	log.Printf("[DEBUG] Called func %s", "getVariableArrayStr")
	var varListStr string
	variableMap := d.Get("wf_variables").(map[string]interface{})

	if variableMap == nil {
		log.Printf("[DEBUG] Does not contain workflow variables in the request!")
		return varListStr
	}
	var variablesStringArr []string
	for var_name, var_value := range variableMap {
		oneVariable := make(map[string]string)
		oneVariable["name"] = var_name
		oneVariable["value"] = var_value.(string)
		oneVariableJsonByte, err := json.Marshal(oneVariable)
		if err != nil {
			fmt.Println(err)
			break
		}
		oneVariableStr := string(oneVariableJsonByte)
		variablesStringArr = append(variablesStringArr, oneVariableStr)
	}

	variablesStringArrLen := len(variablesStringArr)
	for i := 0; i < variablesStringArrLen; i++ {
		if i == 0 {
			varListStr = `"variables":[`
		}
		varListStr = varListStr + variablesStringArr[i]
		if i != variablesStringArrLen-1 {
			varListStr = varListStr + ","
		} else {
			varListStr = varListStr + "]"
		}
	}
	log.Printf("[DEBUG] Variable list str: %s .", varListStr)
	return varListStr
}

func startAWorkflow(ctx context.Context, d *schema.ResourceData, wfKey string, m interface{}) (int, diag.Diagnostics) {
	log.Printf("[DEBUG] Called func %s with workflowKey : %s", "startAWorkflow", wfKey)
	client := m.(Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	//var startWfRequestBody = []byte(fmt.Sprintf(`{
	//        "stepName":"%s"
	//	}`, ""))
	zosmf_workflow_uri_full := fmt.Sprintf("%s/workflow/rest/1.0/workflows", client.HostURL)
	zosmf_start_a_workflow_uri := fmt.Sprintf("%s/%s/operations/start", zosmf_workflow_uri_full, wfKey)
	log.Printf("[DEBUG] Start zOSMF Workflow full path: %s", zosmf_start_a_workflow_uri)
	req, err := http.NewRequest(
		"PUT",
		zosmf_start_a_workflow_uri,
		nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))
	req.AddCookie(&http.Cookie{Name: "LtpaToken2", Value: client.Token, Path: "/", HttpOnly: true})
	log.Printf("[DEBUG] zosmf start a workflow request body: %s", req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	responseCode := resp.StatusCode
	if responseCode == 202 {
		log.Printf("[DEBUG] Start Workflow key: %s, successfully!", wfKey)
	} else {
		log.Printf("[SEVERE] Start Workflow key: %s, failed，responsed code: %d!", wfKey, responseCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] zosmf start a workflow response Body: %s", string(body))

	return responseCode, diags
}

func getPropertiesOfAWorkflow(ctx context.Context, d *schema.ResourceData, wfKey string, m interface{}) (string, diag.Diagnostics) {
	log.Printf("[DEBUG] Called func %s with workflowKey : %s", "getPropertiesOfAWorkflow", wfKey)
	client := m.(Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	//var getWfRequestBody []byte;
	zosmf_workflow_uri_full := fmt.Sprintf("%s/workflow/rest/1.0/workflows", client.HostURL)
	zosmf_get_workflow_properties_uri := fmt.Sprintf("%s/%s", zosmf_workflow_uri_full, wfKey)
	log.Printf("[DEBUG] zOSMF Get Workflow Properties full path: %s", zosmf_get_workflow_properties_uri)
	req, err := http.NewRequest(
		"GET",
		zosmf_get_workflow_properties_uri,
		nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))
	req.AddCookie(&http.Cookie{Name: "LtpaToken2", Value: client.Token, Path: "/", HttpOnly: true})
	log.Printf("[DEBUG] zosmf get a workflow properties request body: %s", req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] zosmf get a workflow properties response Body: %s", string(body))
	responseCode := resp.StatusCode
	wfStatus := "UNKNOWN"
	if responseCode == 200 {
		log.Printf("[DEBUG] Get a workflow properties workflow key: %s, successfully!", wfKey)
		//get "statusName" from json response
		getWfProRes := new(getWfPropertiesResponse)
		if err := json.Unmarshal(body, &getWfProRes); err == nil {
			wfStatus = getWfProRes.StatusName
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("[SEVERE] Get a workflow properties workflow key: %s, failed，responsed code: %d!", wfKey, responseCode)
	}

	return wfStatus, diags
}

func resourceLibertyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "zosmfWorkflowResourceRead")
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] d: %s", d)
	// orderID := d.Id()

	// order, err := c.GetOrder(orderID)
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// orderItems := flattenOrderItems(&order.Items)
	// if err := d.Set("items", orderItems); err != nil {
	// 	return diag.FromErr(err)
	// }

	return diags
}

func resourceLibertyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return zosmfWorkflowResourceRead(ctx, d, m)
}

func resourceLibertyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "resourceLibertyCreate")
	destroyLibertyWfDefPath := d.Get("destroy_liberty_workflow_path").(string)
	destroyLibertyWfVarPath := d.Get("destroy_liberty_variable_file_path").(string)
	workflow_instance_name := d.Get("wf_instance_name").(string) + " stop"

	wfKey, diags := createAWorkflow(ctx, d, destroyLibertyWfDefPath, destroyLibertyWfVarPath, m, workflow_instance_name)
	if wfKey == "" {
		log.Printf("[SEVERE] Create a Workflow failed!")
		return diags
	}
	d.Set("workflow_id", wfKey)
	d.Set("Hostname", "myLocalHost")
	responseCode, diags := startAWorkflow(ctx, d, wfKey, m)
	if responseCode != 202 {
		log.Printf("[SEVERE] Start Workflow key: %s, failed，responsed code: %d!", wfKey, responseCode)
		return diags
	}
	fetchStatusCount := 0
	completeSuccessfully := false
	for {
		time.Sleep(5 * time.Second)
		wfStatus, _ := getPropertiesOfAWorkflow(ctx, d, wfKey, m)
		if wfStatus == "complete" {
			log.Printf("[DEBUG] Workflow workflowKey : %s is complete.", wfKey)
			completeSuccessfully = true
			break
		} else if wfStatus == "automation-in-progress" {
			log.Printf("[DEBUG] Workflow workflowKey : %s is in automation.", wfKey)
		} else {
			log.Printf("[DEBUG] The automation of workflow workflowKey : %s is stop, but the workflow is not complete.", wfKey)
			break
		}
		fetchStatusCount++
		if fetchStatusCount > 3 {
			break
		}
	}

	if completeSuccessfully {
		//	d.SetId(uuid.New().String())
		log.Printf("[DEBUG] libertyResourceDelete ID: %s", d.Id())
		return diags
	} else {
		fmt.Println("The workflow is not complete")
		//	d.SetId(uuid.New().String())
		return diags
	}
}
