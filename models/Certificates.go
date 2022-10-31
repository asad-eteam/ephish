package models

import (
	"fmt"
	"time"

	log "github.com/gophish/gophish/logger"
)

type Certificate struct {
	Id          int64     `json:"-"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Position    string    `json:"position"`
	RId         string    `json:"r_id"`
	CampaignId  int64     `json:"-"`
	CreatedDate time.Time `json:"created_date"`
}

func PostCertificate(rid string) (Certificate, error) {
	ri := rid
	c := Certificate{}
	e := db.Where("r_id=?", rid).Find(&c).Error
	fmt.Println("11111111111")
	fmt.Printf("%+v", c)
	if e != nil {
		fmt.Println(e)
	}
	if c.RId == ri {
		return c, e
	} else {
		r, e := GetResult(ri)
		fmt.Println("rrrrrrrr")
		fmt.Printf("%+v", r)
		if e != nil {
			fmt.Println(e)
		}
		c.RId = r.RId
		c.CampaignId = r.CampaignId
		c.FirstName = r.FirstName
		c.LastName = r.LastName
		c.Email = r.Email
		c.Position = r.Position
		c.CreatedDate = time.Now().UTC().Local()
		err := db.Save(&c).Error
		if err != nil {
			log.Error(err)
		}
		return c, err
	}

}
func GetCertificateById(rid string) (Certificate, error) {
	certificate := Certificate{}
	e := db.Where("r_id=?", rid).Find(&certificate).Error
	if e != nil {
		log.Error(e)
	}
	return certificate, e
}
