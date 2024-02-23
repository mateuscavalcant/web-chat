package models

type User struct {
	ID              int    `json:"id"`
	Icon            []byte `json:"icon"`
	Bio             string `json:"bio" binding:"required, max=150"`
	Name            string `json:"name" binding:"required, min=4,max=32"`
	Email           string `json:"email" binding:"required, email"`
	Password        string `json:"password" binding:"required, min=8, max=16"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}


type UserMessage struct {
	MessageID     int    `json:"post-id"`
	MessageUserID int    `json:"post-user-id"`
	UserID     int    `json:"user-id"`
	Content    string `json:"content"`
	Icon       []byte `json:"icon"`
	IconBase64 string `json:"iconbase64"`
	CreatedBy  string `json:"createdby"`
	Name       string `json:"createdbyname"`
	FollowID int `json:"follow-id"`
	MessageBy int `json:"follow-by"`
	MessageTo  int `json:"follow-to"`
}
