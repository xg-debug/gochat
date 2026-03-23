package response

type UserResponse struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
	Gender    int8   `json:"gender"`
	Phone     string `json:"phone"`
	Location  string `json:"location"`
	Birthday  string `json:"birthday"`
}
