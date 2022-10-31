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

// Pages handles requests for the /api/pages/ endpoint
func (as *Server) Questions(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		ps, err := models.GetQuestions(ctx.Get(r, "user_id").(int64))
		if err != nil {
			log.Error(err)
		}
		JSONResponse(w, ps, http.StatusOK)
	//POST: Create a new page and return it as JSON
	case r.Method == "POST":
		q := models.Question{}
		// Put the request into a page
		err := json.NewDecoder(r.Body).Decode(&q)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Invalid request"}, http.StatusBadRequest)
			return
		}
		q.CreatedDate = time.Now().UTC()
		q.UserId = ctx.Get(r, "user_id").(int64)
		err = models.PostQuestion(&q)
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, q, http.StatusCreated)
	}
}

func (as *Server) Question(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.ParseInt(vars["id"], 0, 64)
	q, err := models.GetPage(id, ctx.Get(r, "user_id").(int64))
	switch {
	case r.Method == "GET":
		JSONResponse(w, q, http.StatusOK)
	case r.Method == "DELETE":
		err = models.DeleteQuestion(id, ctx.Get(r, "user_id").(int64))
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: "Error deleting question"}, http.StatusInternalServerError)
			return
		}
		JSONResponse(w, models.Response{Success: true, Message: "Question Deleted Successfully"}, http.StatusOK)
	}
}
