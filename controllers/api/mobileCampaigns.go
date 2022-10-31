package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

func (as *Server) mobileCampaigns(w http.ResponseWriter, r *http.Request) {
	overview := models.CampaignSummaries{}
	fmt.Println(overview)
	cs := []models.CampaignSummary{}
	if r.Method == "GET" {
		json.NewDecoder(r.Body).Decode(&cs)
		fmt.Println("ur ur ur ur ur : ", &cs)
		w.Header().Set("Content-Type", "application/json")
		appKey := r.Header["Authorization"][0]
		u, err := models.GetUserByAPIKey(appKey)
		uid := u.Id
		c, err := models.GetMobileCampaignSummaries(uid)

		fmt.Println(c)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid User"}, http.StatusBadRequest)
		} else if err == nil {
			JSONResponse(w, models.Response{Success: true, Data: c}, http.StatusAccepted)
		}

	}
}
func (as *Server) mobileCampaignResults(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	w.Header().Set("Content-Type", "application/json")
	appKey := r.Header["Authorization"][0]
	u, err := models.GetUserByAPIKey(appKey)
	uid := u.Id
	cr, err := models.GetCampaignSummary(id, uid)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Campaign not found"}, http.StatusNotFound)
		return
	} else if err == nil {
		JSONResponse(w, models.Response{Success: true, Message: "Campaign found", Data: cr}, http.StatusOK)
	}

}
