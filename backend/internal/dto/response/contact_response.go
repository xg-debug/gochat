package response

type ContactResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Online bool   `json:"online"`
}
