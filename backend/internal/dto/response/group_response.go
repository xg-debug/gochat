package response

type GroupResponse struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
	Notice string `json:"notice"`
	Role   int8   `json:"role"`
}

type GroupMemberResponse struct {
	UserID   int64  `json:"userId"`
	Nickname string `json:"nickname"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Role     int8   `json:"role"`
}
