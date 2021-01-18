package presenter

import "github.com/abcdef-id/go-model/models"

type User struct {
	models.BaseModel
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
}
