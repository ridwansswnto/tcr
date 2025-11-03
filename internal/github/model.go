package github

type WorkflowJobPayload struct {
	Action      string `json:"action"`
	WorkflowJob struct {
		ID              int64    `json:"id"`
		RunID           int64    `json:"run_id"`
		Name            string   `json:"name"`
		Status          string   `json:"status"`
		Conclusion      string   `json:"conclusion"`
		Labels          []string `json:"labels"`
		RunnerGroupName string   `json:"runner_group_name"`
		RunnerName      string   `json:"runner_name"`
	} `json:"workflow_job"`
	Repository struct {
		Name  string `json:"name"`
		Owner struct {
			Login string `json:"login"`
		} `json:"owner"`
	} `json:"repository"`
}
