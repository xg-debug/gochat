package response

type MessageResponse struct {
	ID          string `json:"id"`
	FromID      string `json:"fromId"`
	FromAvatar  string `json:"fromAvatar"`
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
	Time        int64  `json:"time"`
	Status      string `json:"status"`
}
