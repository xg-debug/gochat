package request

type CreateGroupRequest struct {
	Name string `json:"name" binding:"required"`
}

type JoinGroupRequest struct {
	GroupID int64 `json:"groupId" binding:"required"`
}

type UpdateGroupProfileRequest struct {
	GroupID int64  `json:"groupId" binding:"required"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Notice  string `json:"notice"`
}

type GroupIDQuery struct {
	GroupID int64 `form:"groupId" binding:"required"`
}

type KickGroupMemberRequest struct {
	GroupID int64 `json:"groupId" binding:"required"`
	UserID  int64 `json:"userId" binding:"required"`
}

type SetGroupAdminRequest struct {
	GroupID int64  `json:"groupId" binding:"required"`
	UserID  int64  `json:"userId" binding:"required"`
	Action  string `json:"action" binding:"required,oneof=set unset"`
}

type InviteGroupMemberRequest struct {
	GroupID int64 `json:"groupId" binding:"required"`
	UserID  int64 `json:"userId" binding:"required"`
}
