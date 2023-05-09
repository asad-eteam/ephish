package models

import (
	"fmt"
	"time"

	"github.com/gophish/gomail"
	log "github.com/gophish/gophish/logger"
)

// Page contains the fields used for a Page model
type Message struct {
	Id               int64     `json:"id" gorm:"column:id; primary_key:yes"`
	MessageType      string    `json:"message_type"`
	FirstName        string    `json:"first_name"`
	LastName         string    `json:"last_name"`
	OrganizationName string    `json:"organization_name"`
	Email            string    `json:"email"`
	PhoneNumber      string    `json:"phone_number"`
	Message          string    `json:"message"`
	CreatedDate      time.Time `json:"created_date"`
}

func GetMessages() ([]Message, error) {
	qs := []Message{}
	err := db.Order("id desc", true).Find(&qs).Error
	if err != nil {
		fmt.Println(err)
		return qs, err
	}
	fmt.Println(qs)
	return qs, err
}

// PostPage creates a new page in the database.
func PostMessage(m *Message) error {
	// Insert into the DB
	err := db.Save(m).Error
	fmt.Println("##########", err)
	if err != nil {
		log.Error(err)
	}
	// data := db.Delete(m)

	cm := gomail.NewMessage()
	cm.SetHeader("From", "gophish@eteamid.com")
	cm.SetHeader("To", "gophish@eteamid.com")
	cm.SetHeader("Subject", "Hello!")
	cm.SetBody("text/html", " Hello ,<br> "+m.FirstName+" "+m.LastName+" contacted for <b>"+m.MessageType+"</b>"+"<br/><br/>"+
		"<b>Details</b><br>"+m.FirstName+" "+m.LastName+"<br>"+m.Email+"<br>"+m.PhoneNumber)
	// d := gomail.NewPlainDialer("sgp17.siteground.asia", 587, "gophish@eteamid.com", "4k5643g*h)l#")
	// // Send the email to Bob, Cora and Dan.
	// if err := d.DialAndSend(cm); err != nil {
	// 	panic(err)
	// }

	return err
}
func DeleteMessage(id int64) error {

	err := db.Delete(Message{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

func GetMessageById(id int64) ([]Question, error) {
	q := []Question{}
	e := db.Where("id=?", id).Find(&q).Error
	if e != nil {
		log.Error(e)
	}
	return q, e
}
