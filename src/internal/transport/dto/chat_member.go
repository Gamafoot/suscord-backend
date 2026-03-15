package dto

type InviteUserRequest struct {
	UserID uint `json:"user_id" validate:"required,numeric,gt=0"`
}
