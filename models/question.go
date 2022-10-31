package models

import (
	"fmt"
	"time"

	log "github.com/gophish/gophish/logger"
)

// Page contains the fields used for a Page model
type Question struct {
	Id          int64     `json:"id" gorm:"column:id; primary_key:yes"`
	UserId      int64     `json:"-" gorm:"column:user_id"`
	Question    string    `json:"question"`
	Description string    `json:"description"`
	HTML        string    `json:"html" gorm:"column:html"`
	IsPhishing  bool      `json:"is_phishing"`
	CreatedDate time.Time `json:"created_date"`
}

func GetQuestions(uid int64) ([]Question, error) {
	qs := []Question{}
	err := db.Where("user_id=?", uid).Find(&qs).Error
	if err != nil {
		fmt.Println(err)
		return qs, err
	}
	fmt.Println(qs)
	return qs, err
}

// PostPage creates a new page in the database.
func PostQuestion(q *Question) error {
	// Insert into the DB
	err := db.Save(q).Error
	if err != nil {
		log.Error(err)
	}
	return err
}
func DeleteQuestion(id int64, uid int64) error {
	err := db.Where("user_id=?", uid).Delete(Question{Id: id}).Error
	if err != nil {
		log.Error(err)
	}
	return err
}

func GetQuestionById(id int64) ([]Question, error) {
	q := []Question{}
	e := db.Where("id=?", id).Find(&q).Error
	if e != nil {
		log.Error(e)
	}
	return q, e
}
