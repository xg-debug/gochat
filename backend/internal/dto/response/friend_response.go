package response

type UserSearchResult struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	Avatar        string `json:"avatar"`
	IsFriend      bool   `json:"isFriend"`
	Pending       bool   `json:"pending"`
	PendingFromMe bool   `json:"pendingFromMe"`
}

type FriendRequestItem struct {
	ID         int64  `json:"id"`
	FromUserID int64  `json:"fromUserId"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	Avatar     string `json:"avatar"`
	Time       int64  `json:"time"`
}
