package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/gorilla/mux"
)

// Messages handles requests for the /api/messages/ endpoint
func (as *Server) Messages(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ps, err := models.GetMessages()
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ps, http.StatusOK)
	//POST: Create a new message and return it as JSON
	case r.Method == "POST":
		m := models.Message{}
		// Put the request into a message
		err := json.NewDecoder(r.Body).Decode(&m)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		m.CreatedDate = time.Now().UTC()
		err = models.PostMessage(&m)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, m, http.StatusCreated)
	}
}

func (as *Server) Message(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	q, err := models.GetPage(id, ctx.Get(r, "user_id").(int64))
	switch {
	case r.Method == "GET":
		JSONResponse(w, q, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteQuestion(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting Message"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Message Deleted Successfully"}, http.StatusOK)
	}
}
