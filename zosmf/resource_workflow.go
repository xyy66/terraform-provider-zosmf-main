package zosmf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	// "strconv"

	"log"
	"net/http"

	"bytes"

	"io/ioutil"

	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

func zosmfWorkflowResourceSchema() *schema.Resource {
	log.Printf("[DEBUG] Called func %s", "zosmfWorkflowResourceSchema")
	return &schema.Resource{
		CreateContext: zosmfWorkflowResourceCreate,
		ReadContext:   zosmfWorkflowResourceRead,
		UpdateContext: zosmfWorkflowResourceUpdate,
		DeleteContext: zosmfWorkflowResourceDelete,
		Schema: map[string]*schema.Schema{
			"workflow_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"workflow_dir": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"workflow_file_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"system": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func zosmfWorkflowResourceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "zosmfWorkflowResourceCreate")
	client := m.(Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	zosmf_workflow_instance_name := d.Get("instance_name")

	log.Printf("[DEBUG] Instance name     ... %s", zosmf_workflow_instance_name)

	workflow_dir := d.Get("workflow_dir")
	workflow_file_name := d.Get("workflow_file_name")
	system := d.Get("system")

	zosmf_workflow_definition_path_full := fmt.Sprintf("%s%s", workflow_dir, workflow_file_name)

 	var jsonStr = []byte(fmt.Sprintf(`{
            "workflowName":"%s",
            "workflowDefinitionFile":"%s",
            "system":"%s",
            "owner":"%s",
            "assignToOwner":true,
            "variableInputFile":"%s"
		}`, zosmf_workflow_instance_name, zosmf_workflow_definition_path_full, system, client.Username, ""))

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
	log.Printf("[DEBUG] zosmf workflow creation response Body:", string(body))

	workflow := new(Workflow)

	err = json.Unmarshal(body, &workflow)

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("[DEBUG] workflow key: %s", workflow.WorkflowKey)
	d.Set("workflow_id", workflow.WorkflowKey)

	d.SetId(uuid.New().String())

	return diags
}

func zosmfWorkflowResourceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

func zosmfWorkflowResourceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return zosmfWorkflowResourceRead(ctx, d, m)
}

func zosmfWorkflowResourceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// client := m.(Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	log.Printf("[DEBUG] zosmfWorkflowResourceDelete ID: %s", d.Id())
	return diags
}
