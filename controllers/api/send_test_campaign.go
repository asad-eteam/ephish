package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"

	ctx "github.com/gophish/gophish/context"
	// log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	// "github.com/jinzhu/gorm"
	// "github.com/sirupsen/logrus"
)

type Mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

// SendTestEmail sends a test email using the template name
// and Target given.
func (as *Server) SendTestCampaign(w http.ResponseWriter, r *http.Request) {
	tc := &models.TestCampaign{}
	err := json.NewDecoder(r.Body).Decode(&tc)
	if err != nil {
		w.WriteHeader(400)
		fmt.Fprintf(w, "Decode error! please check your JSON formating.")
		return
	}
	s, e := models.GetSMTPByName(tc.SmtpName, ctx.Get(r, "user_id").(int64))
	if e != nil {
		return
	}
	tc.SMTP = s
	//Get email tamplate by given id
	tamplate, e := models.GetTemplateById(tc.TemplateId)
	if e != nil {
		return
	}
        tc.Template = tamplate
        t:=tamplate.Subject
        t=strings.ReplaceAll(t, "{{.FirstName}}", tc.FirstName)
        t=strings.ReplaceAll(t, "{{.LastName}}", tc.LastName)
        t=strings.ReplaceAll(t, "{{.Email}}", tc.Email)
        t=strings.ReplaceAll(t, "{{.Position}}", tc.Position)
    
        tc.Template.Subject = t
	tamplate.HTML = strings.ReplaceAll(tamplate.HTML, "{{.URL}}", "https://whogotphished.com/viewpage/"+strconv.Itoa(int(tc.PageId)))
	
	//Get landing page by given id
	landingPage, e := models.GetPageById(tc.PageId)
	tc.Page = landingPage
	m := tc.Template.HTML
	m = strings.ReplaceAll(m, "{{.FirstName}}", tc.FirstName)
	m = strings.ReplaceAll(m, "{{.LastName}}", tc.LastName)
	m = strings.ReplaceAll(m, "{{.Email}}", tc.Email)
	m = strings.ReplaceAll(m, "{{.Position}}", tc.Position)
	sendCamp(tc.SMTP.Username, tc.Email, tc.SMTP.Username, tc.SMTP.Password, tc.SMTP.FromAddress, tc.SMTP.Host, m, tc.Template.Subject)
	JSONResponse(w, "Test Campaign sent succesfully!", http.StatusCreated)
	return
}
func sendCamp(sender string, tosend string, user string, password string, addr string, host string, msg string, subj string) {
	fmt.Println("*************** smtp smtp smtp *************")
	to := []string{tosend}
	// smtpHost := "smtp.gmail.com"
	hp := host
	h := strings.Split(hp, ":")
	smtpHost := h[0]
	request := Mail{
		Sender:  sender,
		To:      to,
		Subject: subj,
		Body:    string([]byte(msg)),
	}

	msg = BuildMessage(request)
	// Create authentication
	auth := smtp.PlainAuth("", sender, password, smtpHost)

	// Send actual message
	err := smtp.SendMail(host, auth, sender, to, []byte(msg))
	if err != nil {
		log.Fatal(err)
	}
	//**************************
	fmt.Println("Email sent successfully")
	// resp := make(map[strings]strings)

}
func BuildMessage(mail Mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", []byte(mail.Body))

	return msg
}

