package domain

import "time"

type Role string

const (
	RoleUser   Role = "user"
	RoleNormal Role = "normal" // deprecated, use RoleUser instead
	RoleAdmin  Role = "admin"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"size:120;uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Role         Role      `gorm:"type:enum('user','normal','admin');default:'user';not null" json:"rol"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
