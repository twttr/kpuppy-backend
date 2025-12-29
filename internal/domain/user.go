package domain

import "time"

type User struct {
	ID              string    `json:"id"`
	KinopubUsername string    `json:"username"`
	Avatar          *string   `json:"avatar"`
	IsBanned        bool      `json:"-"`
	CreatedAt       time.Time `json:"createdAt"`
}

type UserResponse struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:       u.ID,
		Username: u.KinopubUsername,
		Avatar:   u.Avatar,
	}
}

type ProvisionRequest struct {
	Username string  `json:"username"`
	Avatar   *string `json:"avatar,omitempty"`
}

type ProvisionResponse struct {
	UserID string `json:"userId"`
}

func (r *ProvisionRequest) Validate() error {
	if r.Username == "" {
		return ErrUsernameEmpty
	}
	return nil
}
