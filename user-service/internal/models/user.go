package models

import "time"

type Profile struct {
	UserID    int    `json:"userid"`
	Firstname string `json:"fullname"`
	Lastname  string `json:"lastname"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
	Createdat time.Time
	Updatedat time.Time
}

func NewProfile(userid int, firstname, lastname string, email string, avatar string, createdat time.Time, updatedat time.Time) *Profile {
	return &Profile{
		UserID:    userid,
		Firstname: firstname,
		Lastname:  lastname,
		Email:     email,
		Avatar:    avatar,
		Createdat: createdat,
		Updatedat: updatedat,
	}
}
