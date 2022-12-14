package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB

func (as *Server) contact(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)

	switch {
	case r.Method == "GET":
		m, e := models.GetMessages()

		if e != nil {

			JSONResponse(w, models.Response{Success: false, Message: "not found", Data: e}, http.StatusBadRequest)
		}
		JSONResponse(w, models.Response{Success: true, Message: "Contacts founded", Data: m}, http.StatusOK)
	case r.Method == "DELETE":
		fmt.Println("#######********!!!!!!!! DELETE:::::::::", id)
		err := models.DeleteMessage(id)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting page"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Page Deleted Successfully"}, http.StatusOK)

	}

}
