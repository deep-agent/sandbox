package model

type BashExecRequest struct {
	Command         string            `json:"command" vd:"len($)>0"`
	Cwd             string            `json:"cwd,omitempty"`
	TimeoutMS       int               `json:"timeout_ms,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	RunInBackground bool              `json:"run_in_background,omitempty"`
}

type BashExecResult struct {
	Output     string `json:"output"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error,omitempty"`
	OutputFile string `json:"output_file,omitempty"`
}
