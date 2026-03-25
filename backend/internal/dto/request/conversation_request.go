package request

type SearchConversationsQuery struct {
	Keyword string `form:"keyword"`
}

type MessagesQuery struct {
	ConversationID string `form:"conversationId" binding:"required"`
	Cursor         int64  `form:"cursor"`
	Limit          int    `form:"limit"`
}
