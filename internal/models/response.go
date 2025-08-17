package models

type ErrPayload struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}
type Envelope struct {
	Error    *ErrPayload `json:"error,omitempty"`
	Response any         `json:"response,omitempty"`
	Data     any         `json:"data,omitempty"`
}

type UploadResponse struct {
	Data Data `json:"data"`
}

type Data struct {
	JSON []byte `json:"json"`
	File string `json:"file"`
}

type DocsData struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Mime    string   `json:"mime"`
	File    bool     `json:"file"`
	Public  bool     `json:"public"`
	Created string   `json:"created"`
	Grants  []string `json:"grant"`
}
