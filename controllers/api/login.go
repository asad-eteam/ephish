package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gophish/gophish/models"
)

type ur struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (as *Server) Login(w http.ResponseWriter, r *http.Request) {
	ur := ur{}

	switch {
	case r.Method == "POST":
		json.NewDecoder(r.Body).Decode(&ur)
		fmt.Println("ur ur ur ur ur : ", ur)
		w.Header().Set("Content-Type", "application/json")
		u, err := models.MobileLogin(ur.Username, ur.Password)

		fmt.Println(u)

		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid User"}, http.StatusBadRequest)
		} else if err == nil {
			JSONResponse(w, models.Response{Success: true, Message: "Login Successfully!", Data: u}, http.StatusAccepted)
		}

	}
}
