package models

import "time"

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserProfile struct {
	UserID      string    `json:"user_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Bio         string    `json:"bio"`
	AvatarURL   string    `json:"avatar_url"`
	LastUpdated time.Time `json:"last_updated"`
}

func (u *User) Validate() bool {
	return u.Username != "" && u.Email != ""
}

func (u *User) BeforeCreate() {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
}

func (u *User) BeforeUpdate() {
	u.UpdatedAt = time.Now()
}
