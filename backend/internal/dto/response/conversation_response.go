package response

type ConversationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	LastMessage string `json:"lastMessage"`
	Unread      int    `json:"unread"`
	Online      bool   `json:"online"`
}
