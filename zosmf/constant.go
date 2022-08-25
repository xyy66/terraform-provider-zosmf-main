package zosmf

type Workflow struct {
	WorkflowKey         string `json:"workflowKey"`
	Vendor              string `json:"vendor"`
	WorkflowVersion     string `json:"workflowVersion"`
	WorkflowDescription string `json:"workflowDescription"`
	WorkflowID          string `json:"workflowID"`
}
type Template struct {
	Template_name string `json:"template_name"`
}

var defaultHost string = "xx.xx.xx.xx"
var defaultPort string = "xxxx"
