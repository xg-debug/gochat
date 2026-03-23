package request

type UpdateProfileRequest struct {
	Nickname  *string `json:"nickname"`
	Avatar    *string `json:"avatar"`
	Signature *string `json:"signature"`
	Gender    *int8   `json:"gender"`
	Phone     *string `json:"phone"`
	Location  *string `json:"location"`
	Birthday  *string `json:"birthday"`
}
