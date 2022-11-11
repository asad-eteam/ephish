package models

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/gophish/gophish/logger"
)

type Certificate struct {
	Id          int64  `json:"-"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Email       string `json:"email"`
	Position    string `json:"position"`
	RId         string `json:"r_id"`
	CampaignId  int64  `json:"-"`
	CreatedDate string `json:"created_date"`
	CompanyName string `json:"company_name"`
}

func PostCertificate(rid string) (Certificate, error) {

	c := Certificate{}
	r := Result{}
	// e := db.Where("r_id=?", rid).Find(&c).Error
	// if e != nil {
	// 	fmt.Println(e)
	// }

	// return c, e
	r, e := GetResult(rid)
	if e != nil {
		log.Error("Invalid Id :", e)
	}
	c.RId = r.RId
	c.CampaignId = r.CampaignId
	c.FirstName = r.FirstName
	c.LastName = r.LastName
	c.Email = r.Email
	c.Position = r.Position

	cName, er := GetUser(r.UserId)
	if er != nil {
		fmt.Println(er)
	}
	c.CompanyName = cName.Username

	y := strconv.Itoa(time.Now().Year())
	m := strconv.Itoa(int(time.Now().Month()))
	d := strconv.Itoa(time.Now().Day())
	c.CreatedDate = y + "-" + m + "-" + d
	err := db.Save(&c).Error
	if err != nil {
		log.Error(err)
	}
	return c, err

}
func GetCertificateByRId(rid string) (Certificate, error) {
	c := Certificate{}
	e := db.Where("r_id=?", rid).Find(&c).Error
	if e != nil {
		log.Error(e)
	}
	return c, e
}
