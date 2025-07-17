package models

type Profile struct {
	UserID int    `json:"userid"`
	Email  string `json:"email"`
}

func NewProfile(userid int, email string) *Profile {
	return &Profile{
		UserID: userid,
		Email:  email,
	}
}
