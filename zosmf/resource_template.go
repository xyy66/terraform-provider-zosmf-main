package zosmf

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/tidwall/gjson"

	// "strconv"

	"log"
)

type LibertyAddressAndPort struct {
	libertyAddress string
	libertyPort    string
}

func resourceZosmfLiberty() *schema.Resource {
	log.Printf("[DEBUG] Called func %s", "resourceTemplate")
	return &schema.Resource{
		CreateContext: resourceZosmfLibertyCreate,
		ReadContext:   resourceZosmfLibertyRead,
		UpdateContext: resourceZosmfLibertyUpdate,
		DeleteContext: resourceZosmfLibertyDelete,
		Schema: map[string]*schema.Schema{
			// "template_object_id": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Computed: true,
			// },
			"template_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"software_instance_external_name": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"running_liberty": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			// "template_state": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
			// "tp_action_definition_file": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
			// "tp_workflow_definition_file": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
			// "tp_workflow_variable_input_file": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Optional: true,
			// },
		},
	}
}
func resourceZosmfLibertyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "resourceTemplateCreate")
	client := m.(Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	zosmf_template_name := d.Get("template_name")

	log.Printf("[DEBUG] template name     ... %s", zosmf_template_name)

	zosmf_workflow_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/psc/%s/actions/run", client.HostURL, zosmf_template_name)
	// zosmf_workflow_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/psc/xyy0804/actions/run", client.HostURL)
	log.Printf("[DEBUG] zOSMF liberty full path: %s", zosmf_workflow_uri_full)

	resp := postRequestAndResp(zosmf_workflow_uri_full, m)

	resCode := resp.StatusCode
	log.Printf("[DEBUG] zosmf template creation response code: %d", resCode)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	//object_id := gjson.Get(string(body), "registry-info").Get("object-id")
	object_id := gjson.Get(string(body), "registry-info.object-id")
	d.Set("instance_id", object_id.String())
	/*object_name := gjson.Get(string(body), "registry-info.object-name")
	d.Set("software_instance_name", object_name.String())*/

	log.Printf("[DEBUG] zosmf liberty creation response Body:%s", string(body))

	// After the software instances created and being provisioned
	//time.Sleep(360 * time.Second)
	isSoftwareInstanceBeingProvisioned := false
	fetchStatusCount := 0
	for {
		time.Sleep(10 * time.Second)
		softwareServiceState, diags := getStateOfSoftwareInstance(ctx, d, m)
		if diags != nil {
			log.Printf("[DEBUG] Cannot get the state of software instance!")
			break
		} else if softwareServiceState == "provisioned" {
			log.Printf("[DEBUG] Software instance has been provisioned.")
			isSoftwareInstanceBeingProvisioned = true
			break
		} else if softwareServiceState == "being-provisioned" || softwareServiceState == "being-initialized" {
			log.Printf("[DEBUG] The software instance is being provisioned!")
		} else {
			log.Printf("[DEBUG] The state software instance provisioning failed!")
			break
		}
		fetchStatusCount++
		if fetchStatusCount > 30 {
			break
		}
	}
	// Get the address and port
	if isSoftwareInstanceBeingProvisioned {
		var libertyAddAndPort LibertyAddressAndPort
		libertyAddAndPort, diags = getLibertyAddressAndPortFromSoftwareInstance(ctx, d, m)
		if diags == nil {
			libertyAddress := libertyAddAndPort.libertyAddress
			libertyPort := libertyAddAndPort.libertyPort
			//runningLiberty := map[string]string{"Hostname": "https://172.16.31.56", "Hostport": "9001", "url": "https://172.16.31.56:9001"}
			runningLiberty := fmt.Sprintf("https://%s:%s", libertyAddress, libertyPort)
			d.Set("running_liberty", runningLiberty)
		} else {
			log.Printf("[WARN] Cannot get liberty address and port!")
			runningLiberty := "https://172.16.31.56:9091"
			d.Set("running_liberty", runningLiberty)
		}
		d.SetId(uuid.New().String())
	} else {
		log.Printf("[WARN] The liberty cannot be started!")
		d.SetId(uuid.New().String())
	}
	return diags
}

func getStateOfSoftwareInstance(ctx context.Context, d *schema.ResourceData, m interface{}) (string, diag.Diagnostics) {
	log.Printf("[DEBUG] Called func %s .", "getStateOfSoftwareInstance")
	var state string
	var diags diag.Diagnostics = nil
	if d.Get("instance_id") == nil {
		log.Printf("[DEBUG] The software instance is not defined.")
		diags = diag.Errorf("The software instance is not defined!")
		return state, diags
	}
	softwareInstanceId := (d.Get("instance_id")).(string)
	client := m.(Client)
	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	zosmf_get_software_instance_contents_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/scr/%s", client.HostURL, softwareInstanceId)
	log.Printf("[DEBUG] zOSMF Get software instance state full path: %s", zosmf_get_software_instance_contents_uri_full)

	resp := getRequestAndResp(zosmf_get_software_instance_contents_uri_full, m)

	defer resp.Body.Close()
	responseCode := resp.StatusCode
	log.Printf("[DEBUG] zosmf get software instance state response code: %d", responseCode)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] zosmf get software instance state response Body: %s", string(body))
	if 200 == responseCode {
		state = gjson.Get(string(body), "state").Str
		if d.Get("software_instance_external_name") == nil || d.Get("software_instance_external_name") == "" {
			log.Printf("[DEBUG] Set software_instance_external_name!")
			externalName := gjson.Get(string(body), "external-name").Str
			log.Printf("[DEBUG] zosmf set software instance external name: %s", externalName)
			d.Set("software_instance_external_name", externalName)
		}
		return state, diags
	} else {
		diags = diag.Errorf("Get software instance state failed!")
		return state, diags
	}
}

func getLibertyAddressAndPortFromSoftwareInstance(ctx context.Context, d *schema.ResourceData, m interface{}) (LibertyAddressAndPort, diag.Diagnostics) {
	log.Printf("[DEBUG] Called func %s .", "getLibertyAddressAndPortFromSoftwareInstance")
	var softwareInstanceId string
	var liberyAddAndPort LibertyAddressAndPort
	var diags diag.Diagnostics = nil
	if d.Get("instance_id") == nil {
		log.Printf("[DEBUG] The software instance is not defined.")
		diags = diag.Errorf("The software instance is not defined!")
		return liberyAddAndPort, diags
	}
	softwareInstanceId = (d.Get("instance_id")).(string)
	client := m.(Client)
	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	zosmf_get_software_instance_variables_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/scr/%s/variables", client.HostURL, softwareInstanceId)
	log.Printf("[DEBUG] zOSMF Get software instance variables full path: %s", zosmf_get_software_instance_variables_uri_full)

	resp := getRequestAndResp(zosmf_get_software_instance_variables_uri_full, m)

	defer resp.Body.Close()
	responseCode := resp.StatusCode
	log.Printf("[DEBUG] zosmf get software instance variables response code: %d", responseCode)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Printf("[DEBUG] zosmf get software instance variables response Body: %s", string(body))
	if responseCode == 200 {
		log.Printf("[DEBUG] Get zosmf get software instance variables successfully!")
		var ipAddress string
		var httpsPort string
		softwareInstanceVariables := (gjson.Get(string(body), "variables")).Array()
		for _, v := range softwareInstanceVariables {
			if "IP_ADDRESS" == v.Get("name").Str {
				ipAddress = v.Get("value").Str
				liberyAddAndPort.libertyAddress = ipAddress
			}
			if "HTTPS_PORT" == v.Get("name").Str {
				httpsPort = v.Get("value").Str
				liberyAddAndPort.libertyPort = httpsPort
			}
			if ipAddress != "" && httpsPort != "" {
				break
			}
		}
		log.Printf("[DEBUG] IP_ADDRESS: %s, HTTPS_PORT: %s.\n", ipAddress, httpsPort)
		return liberyAddAndPort, diags
	} else {
		log.Printf("[DEBUG] Get software instance variables failed!")
		diags = diag.Errorf("Get software instance variables failed!")
		return liberyAddAndPort, diags
	}
}

func resourceZosmfLibertyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "zosmfTemplateResourceRead")
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

func resourceZosmfLibertyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceZosmfLibertyRead(ctx, d, m)
}

func resourceZosmfLibertyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// client := m.(Client)
	// Warning or errors can be collected in a slice type

	log.Printf("[DEBUG] Called func %s", "resourceTemplateDelete")
	client := m.(Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	checkSoftwareInstanceResState1, diags := checkSoftwareInstanceExists(ctx, d, m)
	if checkSoftwareInstanceResState1 == "provisioned" {
		zosmf_test_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/scr/%s/actions/deprovision", client.HostURL, d.Get("instance_id"))

		log.Printf("[DEBUG] zOSMF Workflow full path: %s", zosmf_test_uri_full)

		resp := postRequestAndResp(zosmf_test_uri_full, m)

		defer resp.Body.Close()

		resCode := resp.StatusCode
		log.Printf("[DEBUG] deprovision software instance response code: %d", resCode)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("[DEBUG] zosmf start instance response error happens: %s", "err")
		}
		log.Printf("[DEBUG] deprovision software instance response Body:", string(body))
		if resCode == 200 {
			time.Sleep(20 * time.Second)
			isSoftwareInstanceRemoved := false
			fetchStatusCount := 0
			//Check the software instance deleted or not
			for {
				time.Sleep(10 * time.Second)
				checkSoftwareInstanceResState, diags := checkSoftwareInstanceExists(ctx, d, m)
				if diags != nil {
					log.Printf("[DEBUG] Cannot check software instance exists!")
					break
				} else if checkSoftwareInstanceResState == "deprovisioned" {
					log.Printf("[DEBUG] Software instance has been removed.")
					isSoftwareInstanceRemoved = true
					break
				} else if checkSoftwareInstanceResState == "provisioned" {
					log.Printf("[DEBUG] The software instance is being de-provisioned!")
				} else {
					log.Printf("[DEBUG] The state software instance de-provisioning failed!")
				}
				fetchStatusCount++
				if fetchStatusCount > 20 {
					break
				}
			}
			if !isSoftwareInstanceRemoved {
				log.Printf("[WARN] Software instance cannot be removed!")
				diags = diag.Errorf("Software instance cannot be removed!")
			} else {
				log.Printf("[INFO] Software instance has been removed!")
			}
		} else {
			log.Printf("[WARN] Deprovision software instance failed!")
			diags = diag.Errorf("Deprovision software instance failed!")
		}
	} else {
		log.Printf("[DEBUG] The state software instance state is not provisioned !")
		diags = diag.Errorf("The state software instance state is not provisioned !")
	}

	return diags
}

func checkSoftwareInstanceExists(ctx context.Context, d *schema.ResourceData, m interface{}) (string, diag.Diagnostics) {
	log.Printf("[DEBUG] Called func %s .", "checkSoftwareInstanceExists")
	var diags diag.Diagnostics = nil
	if d.Get("instance_id") == nil {
		log.Printf("[DEBUG] The software instance is not defined.")
		diags = diag.Errorf("The software instance is not defined!")
		return "", diags
	}
	softwareInstanceId := (d.Get("instance_id")).(string)
	client := m.(Client)
	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	zosmf_check_software_instance_contents_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/scr/%s", client.HostURL, softwareInstanceId)
	log.Printf("[DEBUG] zOSMF Get software instance state full path: %s", zosmf_check_software_instance_contents_uri_full)

	resp := getRequestAndResp(zosmf_check_software_instance_contents_uri_full, m)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	responseState := gjson.Get(string(body), "state").String()
	log.Printf("[DEBUG] zosmf check software instance exists response state: %s", responseState)
	ioutil.ReadAll(resp.Body)
	return responseState, diags
}

func getFakeRunPublishedSoftwareServiceTemplateResponse() (string, diag.Diagnostics) {
	var diags diag.Diagnostics
	var jsonStr string = `
		{
		"registry-info": {
			"object-name": "QMgr_7",
				"object-id": "c5a8ecdd-db35-466b-aad9-cba0f33bb84b",
				"object-uri": "/zosmf/provisioning/rest/1.0/scr/c5a8ecdd-db35-466b-aad9-cba0f33bb84b"
		},
		"workflow-info": {
			"workflowKey": "ff96459f-27fa-490a-a3e4-4086649c12f3",
				"workflowDescription": "Procedure to provision a MQ for zOS Queue Manager",
				"workflowID": "ProvisionQueueManager",
				"workflowVersion": "1.0.1",
				"vendor": "IBM",
		}
		"system-nickname": "DUMBNODE"
	}
`
	return jsonStr, diags
}

func getRequestAndResp(url string, m interface{}) http.Response {
	client := m.(Client)
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-CSRF-ZOSMF-HEADER", "ZOSMF")
	req.AddCookie(&http.Cookie{Name: "LtpaToken2", Value: client.Token, Path: "/", HttpOnly: true})
	log.Printf("[DEBUG] zosmf create liberty request body: %s", req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return *resp
}
func postRequestAndResp(url string, m interface{}) http.Response {
	client := m.(Client)
	req, err := http.NewRequest(
		"POST",
		url,
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-CSRF-ZOSMF-HEADER", "ZOSMF")
	req.AddCookie(&http.Cookie{Name: "LtpaToken2", Value: client.Token, Path: "/", HttpOnly: true})
	log.Printf("[DEBUG] zosmf create liberty request body: %s", req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return *resp
}
