package api

import (
	"fmt"
	"net/http"

	"github.com/gophish/gophish/models"
)

func (as *Server) Quiz(w http.ResponseWriter, r *http.Request) {
	q, e := models.GetQuestionById(1)
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(q)
	JSONResponse(w, q, http.StatusCreated)
}
