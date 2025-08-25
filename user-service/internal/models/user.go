package models

import (
	"time"

	"github.com/google/uuid"
)

type Profile struct {
	UserID    uuid.UUID `json:"userid"`
	Firstname string    `json:"fullname"`
	Lastname  string    `json:"lastname"`
	Email     string    `json:"email"`
	//Avatar    string `json:"avatar"`
	Createdat time.Time
	Updatedat time.Time
}

func NewProfile(userid uuid.UUID, firstname, lastname string, email string, createdat time.Time, updatedat time.Time) *Profile {
	return &Profile{
		UserID:    userid,
		Firstname: firstname,
		Lastname:  lastname,
		Email:     email,
		//Avatar:    avatar,
		Createdat: createdat,
		Updatedat: updatedat,
	}
}
