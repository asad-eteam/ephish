package models

import "fmt"

// var db *gorm.DB
type Contacts struct {
	UserId     int64  `json:"userId"`
	CampaignId int64  `json:"campaignId"`
	Message    string `json:"message"`
}

func GetContacts() (Contacts, error) {
	c := Contacts{}
	query := db.Table("contacts").Scan(&c)
	fmt.Println(query)
	return c, nil
}
