package request

type SearchUserQuery struct {
	Keyword string `form:"keyword" binding:"required"`
}

type SendFriendRequest struct {
	ToUserID int64 `json:"toUserId" binding:"required"`
}

type FriendActionRequest struct {
	FriendID int64 `json:"friendId" binding:"required"`
}

type HandleFriendRequest struct {
	RequestID int64  `json:"requestId" binding:"required"`
	Action    string `json:"action" binding:"required,oneof=accept reject"`
}
