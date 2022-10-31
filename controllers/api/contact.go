package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gophish/gophish/models"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func (as *Server) contact(w http.ResponseWriter, r *http.Request) {
	contact := models.Contacts{}
	switch {
	case r.Method == "GET":
		c, e := models.GetContacts()
		fmt.Println(c)
		if e != nil {

			JSONResponse(w, models.Response{Success: false, Message: "not found", Data: e}, http.StatusBadRequest)
		}
		JSONResponse(w, models.Response{Success: true, Message: "Contacts founded", Data: c}, http.StatusOK)

	case r.Method == "POST":
		json.NewDecoder(r.Body).Decode(&contact)
		appKey := r.Header["Authorization"][0]
		u, err := models.GetUserByAPIKey(appKey)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid User"}, http.StatusBadRequest)
		}
		uid := u.Id

		c, e := models.AddMessag(uid, contact.CampaignId, contact.Message)
		fmt.Println(c)
		if e != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Something went wrong!"}, http.StatusBadRequest)
		}
		JSONResponse(w, models.Response{Success: true, Message: "Request Submit Successfully!"}, http.StatusBadRequest)
	}

}
