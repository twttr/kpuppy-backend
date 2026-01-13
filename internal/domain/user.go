package domain

import "time"

type User struct {
	ID          string    `json:"id"`
	UserHash    string    `json:"-"`
	DisplayName string    `json:"displayName"`
	Avatar      *string   `json:"avatar,omitempty"`
	IsBanned    bool      `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
}

type UserResponse struct {
	ID          string  `json:"id"`
	DisplayName string  `json:"displayName"`
	Avatar      *string `json:"avatar,omitempty"`
}

func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		Avatar:      u.Avatar,
	}
}

type ProvisionRequest struct {
	UserHash string  `json:"userHash"`
	Avatar   *string `json:"avatar,omitempty"`
}

type ProvisionResponse struct {
	UserID      string `json:"userId"`
	DisplayName string `json:"displayName"`
}

func (r *ProvisionRequest) Validate() error {
	if r.UserHash == "" {
		return ErrUserHashEmpty
	}
	if len(r.UserHash) != 64 {
		return ErrInvalidUserHash
	}
	return nil
}
