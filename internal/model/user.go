package model

type User struct {
	BaseModel
	Username       string
	PasswordHashed string
	Email          string
}
