package model

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type SandboxContext struct {
	HomeDir   string `json:"home_dir"`
	Workspace string `json:"workspace"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}
