package dao

type UserDTO struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"rol"`
}
