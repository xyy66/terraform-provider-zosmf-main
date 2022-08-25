package zosmf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

	"errors"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"zosmf_username": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZOSMF_USERNAME", nil),
			},
			"zosmf_password": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZOSMF_PASSWORD", nil),
			},
			"zosmf_url": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ZOSMF_URL", nil),
			},
			"allow_unverified_ssl": &schema.Schema{
				Type:      schema.TypeBool,
				Optional:  true,
				Sensitive: true,
				Default:   true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"zosmf_workflow":         zosmfWorkflowResourceSchema(),
			"zosmf_liberty":          resourceLiberty(),
			"zosmf_resource_liberty": resourceZosmfLiberty(), 
			// "zosmf_instance": resourceInstance(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

type Client struct {
	HostURL    string // zosmf host url, ex. https://zosmf.com:443/zosmf/
	HTTPClient *http.Client
	Token      string // Ltpatoken2 token after a successful login to zosmf
	Username   string
	ClientIp   string
}

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Retrieve client IP
	client_ip, err := externalIP()
	if err != nil {
		fmt.Println(err)
	}
	log.Printf("[DEBUG] Client IP: %s", client_ip)

	// zOSMF login client
	skipTLS := d.Get("allow_unverified_ssl").(bool)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipTLS},
	}

	client := Client{
		HTTPClient: &http.Client{Transport: tr, Timeout: 30 * time.Second},
		// Default Hashicups URL
		HostURL:  d.Get("zosmf_url").(string),
		ClientIp: client_ip,
	}

	zosmf_username := d.Get("zosmf_username").(string)
	zosmf_password := d.Get("zosmf_password").(string)

	data := url.Values{}
	data.Add("requestType", "Login")
	data.Add("username", zosmf_username)
	data.Add("password", zosmf_password)

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("%s/LoginServlet", client.HostURL),
		bytes.NewBufferString(data.Encode()))

	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value") // important!
	req.Header.Add("Host", client.HostURL)
	req.Header.Add("Origin", client.HostURL)
	req.Header.Add("Referer", fmt.Sprintf("%s/LogOnPanel.jsp", client.HostURL))

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		log.Fatal(err)
		return nil, diag.FromErr(errors.New("Login request to zOSMF failed"))
	}
	defer resp.Body.Close()

	var zosmf_login_cookie = ""
	for _, cookie := range resp.Cookies() {
		log.Println("Found a cookie named:", cookie.Name)
		if cookie.Name == "LtpaToken2" {
			zosmf_login_cookie = cookie.Value
		}
	}

	if zosmf_login_cookie == "" {
		return nil, diag.FromErr(errors.New("zOSMF login failed, no LtpaToken2 retrieved, check credential"))
	}

	log.Printf("[DEBUG] zosmf login cookie: %s", zosmf_login_cookie)

	client.Token = zosmf_login_cookie
	client.Username = zosmf_username
	return client, diags
}
