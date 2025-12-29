package domain

import "time"

type Content struct {
	ID            string    `json:"id"`
	KinopubItemID int64     `json:"kinopubItemId"`
	CreatedAt     time.Time `json:"createdAt"`
}
