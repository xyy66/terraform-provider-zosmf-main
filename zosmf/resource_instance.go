package zosmf

import (
	"bytes"
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	// "strconv"

	"log"
	"net/http"

	"io/ioutil"

	"fmt"

	"github.com/google/uuid"
)

func resourceInstance() *schema.Resource {
	log.Printf("[DEBUG] Called func %s", "resourceInstance")
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		UpdateContext: resourceInstanceUpdate,
		DeleteContext: resourceInstanceDelete,
		Schema: map[string]*schema.Schema{
			"instance_object_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			// "template_name": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
			// "tp_domain_name": &schema.Schema{
			// 	Type:     schema.TypeString,
			// 	Required: true,
			// },
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

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "resourceInstanceCreate")
	client := m.(Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	log.Printf("[DEBUG] zosmf login token : %v", client.Token)

	object_id := d.Get("instance_object_id")

	var jsonStr = []byte(fmt.Sprintf(`{
		}`))
	// var jsonStrrun = []byte(fmt.Sprintf(`{
	// 	}`))
	//var hostPort = "443"
	zosmf_test_uri_full := fmt.Sprintf("%s/provisioning/rest/1.0/scr/%s/actions/start", client.HostURL, object_id)
	// var zosmf_test_uri_full = "https://172.16.31.56:443/zosmf/provisioning/rest/1.0/scr/bf52d162-c0e0-4ad4-9045-e04f2706553a/actions/start"
	// zosmf_instance_test_url := fmt.Sprintf("%s/provisioning/rest/1.0/src/bf52d162-c0e0-4ad4-9045-e04f2706553a/actions/start", client.HostURL)

	log.Printf("[DEBUG] zOSMF Workflow full path: %s", zosmf_test_uri_full)
	req, err := http.NewRequest(
		"POST",
		zosmf_test_uri_full,
		bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))
	req.Header.Add("X-CSRF-ZOSMF-HEADER", "ZOSMF")
	req.Header.Add("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "LtpaToken2", Value: client.Token, Path: "/", HttpOnly: true})
	log.Printf("[DEBUG] zosmf start instance request body: %s", req)

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Printf("[DEBUG] zosmf start instance request error happens: %s", "err")
		log.Fatal(err)
	}
	defer resp.Body.Close()

	resCode := resp.StatusCode
	log.Printf("[DEBUG] zosmf instance creation response code: %d", resCode)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[DEBUG] zosmf start instance response error happens: %s", "err")
	}
	log.Printf("[DEBUG] zosmf instance creation response Body:", string(body))

	d.SetId(uuid.New().String())

	return diags
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Called func %s", "zosmftemplateRead")
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
func resourceInstanceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return zosmfWorkflowResourceRead(ctx, d, m)
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// client := m.(Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	log.Printf("[DEBUG] zosmfWorkflowResourceDelete ID: %s", d.Id())
	return diags
}
