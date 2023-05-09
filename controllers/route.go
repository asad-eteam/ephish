package controllers

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/gophish/gomail"
	"github.com/gophish/gophish/auth"
	"github.com/gophish/gophish/config"
	ctx "github.com/gophish/gophish/context"
	"github.com/gophish/gophish/controllers/api"
	log "github.com/gophish/gophish/logger"
	mid "github.com/gophish/gophish/middleware"
	"github.com/gophish/gophish/middleware/ratelimit"
	"github.com/gophish/gophish/models"
	"github.com/gophish/gophish/util"
	"github.com/gophish/gophish/worker"
	"github.com/gorilla/csrf"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jordan-wright/unindexed"
	"github.com/jung-kurt/gofpdf"
	"github.com/jung-kurt/gofpdf/contrib/httpimg"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

type mail struct {
	Sender  string
	To      []string
	Subject string
	Body    string
}

// AdminServerOption is a functional option that is used to configure the
// admin server
type AdminServerOption func(*AdminServer)

// AdminServer is an HTTP server that implements the administrative Gophish
// handlers, including the dashboard and REST API.
type AdminServer struct {
	server  *http.Server
	worker  worker.Worker
	config  config.AdminServer
	limiter *ratelimit.PostLimiter
}

var defaultTLSConfig = &tls.Config{
	PreferServerCipherSuites: true,
	CurvePreferences: []tls.CurveID{
		tls.X25519,
		tls.CurveP256,
	},
	MinVersion: tls.VersionTLS12,
	CipherSuites: []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,

		// Kept for backwards compatibility with some clients
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	},
}

// WithWorker is an option that sets the background worker.
func WithWorker(w worker.Worker) AdminServerOption {
	return func(as *AdminServer) {
		as.worker = w
	}
}

// NewAdminServer returns a new instance of the AdminServer with the
// provided config and options applied.
func NewAdminServer(config config.AdminServer, options ...AdminServerOption) *AdminServer {
	defaultWorker, _ := worker.New()
	defaultServer := &http.Server{
		ReadTimeout: 10 * time.Second,
		Addr:        config.ListenURL,
	}
	defaultLimiter := ratelimit.NewPostLimiter()
	as := &AdminServer{
		worker:  defaultWorker,
		server:  defaultServer,
		limiter: defaultLimiter,
		config:  config,
	}
	for _, opt := range options {
		opt(as)
	}
	as.registerRoutes()
	return as
}

// Start launches the admin server, listening on the configured address.
func (as *AdminServer) Start() {
	if as.worker != nil {
		go as.worker.Start()
	}
	if as.config.UseTLS {
		// Only support TLS 1.2 and above - ref #1691, #1689
		as.server.TLSConfig = defaultTLSConfig
		err := util.CheckAndCreateSSL(as.config.CertPath, as.config.KeyPath)
		if err != nil {
			log.Fatal(err)
		}
		log.Infof("Starting admin server at https://%s", as.config.ListenURL)
		log.Fatal(as.server.ListenAndServeTLS(as.config.CertPath, as.config.KeyPath))
	}
	// If TLS isn't configured, just listen on HTTP
	log.Infof("Starting admin server at http://%s", as.config.ListenURL)
	log.Fatal(as.server.ListenAndServe())
}

// Shutdown attempts to gracefully shutdown the server.
func (as *AdminServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	return as.server.Shutdown(ctx)
}

// SetupAdminRoutes creates the routes for handling requests to the web interface.
// This function returns an http.Handler to be used in http.ListenAndServe().
func (as *AdminServer) registerRoutes() {

	router := mux.NewRouter()
	fs := http.FileServer(http.Dir("./static/"))
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))
	router.HandleFunc("/", as.Home)
	router.HandleFunc("/dashboard", mid.Use(as.Base, mid.RequireLogin))
	router.HandleFunc("/login", mid.Use(as.Login, as.limiter.Limit))
	router.HandleFunc("/sso", mid.Use(as.Sso, as.limiter.Limit))
	router.HandleFunc("/logout", mid.Use(as.Logout, mid.RequireLogin))
	router.HandleFunc("/reset_password", mid.Use(as.ResetPassword, mid.RequireLogin))
	router.HandleFunc("/campaigns", mid.Use(as.Campaigns, mid.RequireLogin))
	router.HandleFunc("/campaigns/{id:[0-9]+}", mid.Use(as.CampaignID, mid.RequireLogin))
	router.HandleFunc("/templates", mid.Use(as.Templates, mid.RequireLogin))
	router.HandleFunc("/groups", mid.Use(as.Groups, mid.RequireLogin))
	router.HandleFunc("/landing_pages", mid.Use(as.LandingPages, mid.RequireLogin))
	router.HandleFunc("/sending_profiles", mid.Use(as.SendingProfiles, mid.RequireLogin))
	router.HandleFunc("/settings", mid.Use(as.Settings, mid.RequireLogin))
	router.HandleFunc("/users", mid.Use(as.UserManagement, mid.RequirePermission(models.PermissionModifySystem), mid.RequireLogin))
	router.HandleFunc("/webhooks", mid.Use(as.Webhooks, mid.RequirePermission(models.PermissionModifySystem), mid.RequireLogin))
	router.HandleFunc("/impersonate", mid.Use(as.Impersonate, mid.RequirePermission(models.PermissionModifySystem), mid.RequireLogin))
	router.HandleFunc("/contactus", mid.Use(as.Contacts, mid.RequireLogin))
	router.HandleFunc("/questions", mid.Use(as.Questions, mid.RequireLogin))
	router.HandleFunc("/report", as.Report)
	router.HandleFunc("/test", as.Test)
	router.HandleFunc("/quiz", as.quiz)
	router.HandleFunc("/certificate", as.CreateCertificate)
	router.HandleFunc("/message", as.Message).Methods("POST")
	router.HandleFunc("/viewpage/{id:[0-9]+}", mid.Use(ViewPage))

	// Create the API routes
	api := api.NewServer(
		api.WithWorker(as.worker),
		api.WithLimiter(as.limiter),
	)
	router.PathPrefix("/api/").Handler(api)
	// Setup static file serving
	router.PathPrefix("/").Handler(http.FileServer(unindexed.Dir("./static/")))

	// Setup CSRF Protection
	csrfKey := []byte(as.config.CSRFKey)
	if len(csrfKey) == 0 {
		csrfKey = []byte(auth.GenerateSecureKey(auth.APIKeyLength))
	}
	csrfHandler := csrf.Protect(csrfKey,
		csrf.FieldName("csrf_token"),
		csrf.Secure(as.config.UseTLS))
	adminHandler := csrfHandler(router)
	adminHandler = mid.Use(adminHandler.ServeHTTP, mid.CSRFExceptions, mid.GetContext, mid.ApplySecurityHeaders)

	// Setup GZIP compression
	gzipWrapper, _ := gziphandler.NewGzipLevelHandler(gzip.BestCompression)
	adminHandler = gzipWrapper(adminHandler)

	// Respect X-Forwarded-For and X-Real-IP headers in case we're behind a
	// reverse proxy.
	adminHandler = handlers.ProxyHeaders(adminHandler)

	// Setup logging
	adminHandler = handlers.CombinedLoggingHandler(log.Writer(), adminHandler)
	as.server.Handler = adminHandler
}

type templateParams struct {
	Title        string
	Flashes      []interface{}
	User         models.User
	Token        string
	Version      string
	ModifySystem bool
}

// newTemplateParams returns the default template parameters for a user and
// the CSRF token.
func newTemplateParams(r *http.Request) templateParams {
	user := ctx.Get(r, "user").(models.User)
	session := ctx.Get(r, "session").(*sessions.Session)
	modifySystem, _ := user.HasPermission(models.PermissionModifySystem)
	return templateParams{
		Token:        csrf.Token(r),
		User:         user,
		ModifySystem: modifySystem,
		Version:      config.Version,
		Flashes:      session.Flashes(),
	}
}

// Base handles the default path and template execution
func (as *AdminServer) Base(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Dashboard"
	getTemplate(w, "dashboard").ExecuteTemplate(w, "base", params)
}

// Campaigns handles the default path and template execution
func (as *AdminServer) Campaigns(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Campaigns"
	getTemplate(w, "campaigns").ExecuteTemplate(w, "base", params)

}

// CampaignID handles the default path and template execution
func (as *AdminServer) CampaignID(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Campaign Results"
	getTemplate(w, "campaign_results").ExecuteTemplate(w, "base", params)
}

// Templates handles the default path and template execution
func (as *AdminServer) Templates(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Email Templates"
	getTemplate(w, "templates").ExecuteTemplate(w, "base", params)
}

// Groups handles the default path and template execution
func (as *AdminServer) Groups(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Users & Groups"
	getTemplate(w, "groups").ExecuteTemplate(w, "base", params)
}

// LandingPages handles the default path and template execution
func (as *AdminServer) LandingPages(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Landing Pages"
	getTemplate(w, "landing_pages").ExecuteTemplate(w, "base", params)
}

// SendingProfiles handles the default path and template execution
func (as *AdminServer) SendingProfiles(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Sending Profiles"
	getTemplate(w, "sending_profiles").ExecuteTemplate(w, "base", params)
}
func (as *AdminServer) Report(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./templates/report.html")
}
func (as *AdminServer) Test(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%+v", r)
	t, _ := template.ParseFiles("./templates/test/index.html")
	t.Execute(w, nil)
}
func (as *AdminServer) quiz(w http.ResponseWriter, r *http.Request) {
	rid := r.URL.Query().Get("rid")
	id := r.URL.Query().Get("id")
	result, e := models.GetResult(rid)
	if e != nil {
		fmt.Println(e)
	}
	if result.RId != rid {
		t, _ := template.ParseFiles("./templates/quiz/message.html")
		t.Execute(w, nil)
		return
	}
	t, _ := template.ParseFiles("./templates/quiz/" + id + ".html")
	t.Execute(w, nil)

}
func (as *AdminServer) Message(w http.ResponseWriter, r *http.Request) {
	u, e := models.GetUser(1)
	if e != nil {

	}
	csrf.Protect(
		[]byte(u.ApiKey),
		csrf.HttpOnly(false),
		csrf.Secure(false),
	)
	csrf.Token(r)
	// params := struct {
	// 	User    models.User
	// 	Title   string
	// 	Flashes []interface{}
	// 	Token   string
	// }{Title: "Login", Token: csrf.Token(r)}
	fmt.Println("*************&&&&&&&&&&&&&&&&&############")
}
func (as *AdminServer) CreateCertificate(w http.ResponseWriter, r *http.Request) {

	rid := r.URL.Query().Get("rid")
	_, e := models.GetResult(rid)

	if e != nil {

		log.Error("Invalid Id :", e)
		return
	}

	certificate, e := models.PostCertificate(rid)

	if e != nil {
		fmt.Println(e)
	}
	pdf := gofpdf.New("p", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	if as.config.UseTLS {

	}
	url := "https://eteamid.com/gophish/certificate.jpg"
	httpimg.Register(pdf, url, "")
	pdf.Image(url, 8, 10, 195, 0, false, "", 0, "")
	htmlStr :=
		`<br><br><br><br><br><br><br><br><br><center><b><i>Certificate of Completion <br>Congratulation</i></b></center>` +
			`<center><i>` + certificate.FirstName + ` ` + certificate.LastName + `</i></center>` +
			`<center><i>` + certificate.Position + `</i></center><br>` +
			`<center><i>You have successfully completed following course</i></center>` +
			`<center><i>Phishing Awareness Challenge</i></center><br><br>` +
			`<center><i>Company Name: ` + certificate.CompanyName + `</i></center>` +
			`<center><i>Certificate ID: ` + rid + `</i></center>` +
			`<center><i>Issue Date: ` + certificate.CreatedDate + `</i></center><br><br>`

	html := pdf.HTMLBasicNew()
	_, lineHt := pdf.GetFontSize()
	html.Write(lineHt, htmlStr)

	err := pdf.OutputFileAndClose("./static/certificates/" + rid + ".pdf")
	if err != nil {
		fmt.Println("eeeeeee", err)
	}
	result, e := models.GetResult(rid)
	sm, se := models.GetSMTPUserId(result.UserId)
	if se != nil {

	}
	to := []string{certificate.Email}
	sender := sm.Username
	password := sm.Password
	h := strings.Split(sm.Host, ":")[0]
	port, err := strconv.Atoi(strings.Split(sm.Host, ":")[1])
	if err != nil {
		log.Error(err)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", to...)
	m.SetHeader("Subject", "Hello!")
	m.SetBody("text/html", " Hello "+certificate.FirstName+",<br> Congratulations on completing the <b>&ldquo;Phishing Awareness Certification&rdquo;</b>.<br>Your certificate is attached with this email and also available on this link:"+
		"  <a download href="+"https://whogotphished.com/static/certificates/"+rid+".pdf"+">Click Here</a><br><br><br>Thank you!<br><br>From:<br> <a href='https://whogotphished.com'>WhoGotPhished.com</a> ")
	m.Attach("./static/certificates/" + rid + ".pdf")
	fmt.Println("111111111111111")
	d := gomail.NewPlainDialer(h, port, sender, password)
	fmt.Println("222222222222222")
	// Send the email to Bob, Cora and Dan.
	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}

	fmt.Println("Email sent successfully")
	t, _ := template.ParseFiles("./templates/quiz/CertificateSent.html")
	Sent := certificate.Email
	t.Execute(w, Sent)

}
func (as *AdminServer) Home(w http.ResponseWriter, r *http.Request) {

	// t, _ := template.ParseFiles("./templates/website/home.html")
	// t.Execute(w, params)

	params := struct {
		Title    string
		Flashes  []interface{}
		Token    string
		Messages string
	}{Title: "Guest", Messages: "", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)

	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		t, _ := template.ParseFiles("./templates/website/home.html")
		t.Execute(w, params)
	case r.Method == "POST":
		params.Messages = "success"
		recaptchaResponse := r.FormValue("g-recaptcha-response")
		captcha, _ := recaptcha.NewReCAPTCHA("6LfsVQUkAAAAAIrxI6-o5JFI9lQUntVTciuGmmsK", recaptcha.V2, 10*time.Second) // for v2 API get your secret from https://www.google.com/recaptcha/admin
		err := captcha.Verify(recaptchaResponse)
		if err != nil {
			params.Messages = "Invalid Captcha"
		}
		// proceed
		m := models.Message{}
		m.MessageType = r.FormValue("type")
		m.FirstName = r.FormValue("firstname")
		m.LastName = r.FormValue("lastname")
		m.OrganizationName = r.FormValue("organizationname")
		m.PhoneNumber = r.FormValue("phonenumber")
		m.Email = r.FormValue("email")
		m.Message = r.FormValue("message")
		m.CreatedDate = time.Now()

		data := models.PostMessage(&m)

		if data != nil {
			params.Messages = "error"
		}
		t, e := template.ParseFiles("./templates/website/home.html")
		if e != nil {

		}
		t.Execute(w, params)
	}
}

// Settings handles the changing of settings
func (as *AdminServer) Settings(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Method == "GET":
		params := newTemplateParams(r)
		params.Title = "Settings"
		session := ctx.Get(r, "session").(*sessions.Session)
		session.Save(r, w)
		getTemplate(w, "settings").ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		u := ctx.Get(r, "user").(models.User)
		currentPw := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_new_password")
		// Check the current password
		err := auth.ValidatePassword(currentPw, u.Hash)
		msg := models.Response{Success: true, Message: "Settings Updated Successfully"}
		if err != nil {
			msg.Message = err.Error()
			msg.Success = false
			api.JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
		newHash, err := auth.ValidatePasswordChange(u.Hash, newPassword, confirmPassword)
		if err != nil {
			msg.Message = err.Error()
			msg.Success = false
			api.JSONResponse(w, msg, http.StatusBadRequest)
			return
		}
		u.Hash = string(newHash)
		if err = models.PutUser(&u); err != nil {
			msg.Message = err.Error()
			msg.Success = false
			api.JSONResponse(w, msg, http.StatusInternalServerError)
			return
		}
		api.JSONResponse(w, msg, http.StatusOK)
	}
}

// UserManagement is an admin-only handler that allows for the registration
// and management of user accounts within Gophish.
func (as *AdminServer) UserManagement(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "User Management"
	getTemplate(w, "users").ExecuteTemplate(w, "base", params)
}

func (as *AdminServer) nextOrIndex(w http.ResponseWriter, r *http.Request) {
	next := "/dashboard"
	url, err := url.Parse(r.FormValue("next"))
	if err == nil {
		path := url.Path
		if path != "" {
			next = path
		}
	}
	http.Redirect(w, r, next, http.StatusFound)
}

func (as *AdminServer) handleInvalidLogin(w http.ResponseWriter, r *http.Request, message string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	Flash(w, r, "danger", message)
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Login", Token: csrf.Token(r)}
	params.Flashes = session.Flashes()
	session.Save(r, w)
	templates := template.New("template")
	_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
	if err != nil {
		log.Error(err)
	}
	// w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	template.Must(templates, err).ExecuteTemplate(w, "base", params)
}

// Webhooks is an admin-only handler that handles webhooks
func (as *AdminServer) Webhooks(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Webhooks"
	getTemplate(w, "webhooks").ExecuteTemplate(w, "base", params)
}

// Impersonate allows an admin to login to a user account without needing the password
func (as *AdminServer) Impersonate(w http.ResponseWriter, r *http.Request) {

	if r.Method == "POST" {
		username := r.FormValue("username")
		u, err := models.GetUserByUsername(username)
		if err != nil {
			log.Error(err)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		session := ctx.Get(r, "session").(*sessions.Session)
		session.Values["id"] = u.Id
		session.Save(r, w)
	}
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// Login handles the authentication flow for a user. If credentials are valid,
// a session is created
func (as *AdminServer) Login(w http.ResponseWriter, r *http.Request) {
	params := struct {
		User    models.User
		Title   string
		Flashes []interface{}
		Token   string
	}{Title: "Login", Token: csrf.Token(r)}
	session := ctx.Get(r, "session").(*sessions.Session)

	switch {
	case r.Method == "GET":
		params.Flashes = session.Flashes()
		session.Save(r, w)
		templates := template.New("template")
		_, err := templates.ParseFiles("templates/login.html", "templates/flashes.html")
		if err != nil {
			log.Error(err)
		}
		template.Must(templates, err).ExecuteTemplate(w, "base", params)
	case r.Method == "POST":
		// Find the user with the provided username
		username, password := r.FormValue("username"), r.FormValue("password")
		u, err := models.GetUserByUsername(username)
		if err != nil {
			log.Error(err)
			as.handleInvalidLogin(w, r, "Invalid Username/Password")
			return
		}
		// Validate the user's password
		err = auth.ValidatePassword(password, u.Hash)
		if err != nil {
			log.Error(err)
			as.handleInvalidLogin(w, r, "Invalid Username/Password")
			return
		}
		if u.AccountLocked {
			as.handleInvalidLogin(w, r, "Account Locked")
			return
		}
		u.LastLogin = time.Now().UTC()
		err = models.PutUser(&u)
		if err != nil {
			log.Error(err)
		}
		// If we've logged in, save the session and redirect to the dashboard
		session.Values["id"] = u.Id
		session.Save(r, w)
		as.nextOrIndex(w, r)
	}
}

// Logout destroys the current user session
func (as *AdminServer) Logout(w http.ResponseWriter, r *http.Request) {
	session := ctx.Get(r, "session").(*sessions.Session)
	delete(session.Values, "id")
	Flash(w, r, "success", "You have successfully logged out")
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// ResetPassword handles the password reset flow when a password change is
// required either by the Gophish system or an administrator.
//
// This handler is meant to be used when a user is required to reset their
// password, not just when they want to.
//
// This is an important distinction since in this handler we don't require
// the user to re-enter their current password, as opposed to the flow
// through the settings handler.
//
// To that end, if the user doesn't require a password change, we will
// redirect them to the settings page.
func (as *AdminServer) ResetPassword(w http.ResponseWriter, r *http.Request) {
	u := ctx.Get(r, "user").(models.User)
	session := ctx.Get(r, "session").(*sessions.Session)
	if !u.PasswordChangeRequired {
		Flash(w, r, "info", "Please reset your password through the settings page")
		session.Save(r, w)
		http.Redirect(w, r, "/settings", http.StatusTemporaryRedirect)
		return
	}
	params := newTemplateParams(r)
	params.Title = "Reset Password"
	switch {
	case r.Method == http.MethodGet:
		params.Flashes = session.Flashes()
		session.Save(r, w)
		getTemplate(w, "reset_password").ExecuteTemplate(w, "base", params)
		return
	case r.Method == http.MethodPost:
		newPassword := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		newHash, err := auth.ValidatePasswordChange(u.Hash, newPassword, confirmPassword)
		if err != nil {
			Flash(w, r, "danger", err.Error())
			params.Flashes = session.Flashes()
			session.Save(r, w)
			w.WriteHeader(http.StatusBadRequest)
			getTemplate(w, "reset_password").ExecuteTemplate(w, "base", params)
			return
		}
		u.PasswordChangeRequired = false
		u.Hash = newHash
		if err = models.PutUser(&u); err != nil {
			Flash(w, r, "danger", err.Error())
			params.Flashes = session.Flashes()
			session.Save(r, w)
			w.WriteHeader(http.StatusInternalServerError)
			getTemplate(w, "reset_password").ExecuteTemplate(w, "base", params)
			return
		}
		// TODO: We probably want to flash a message here that the password was
		// changed successfully. The problem is that when the user resets their
		// password on first use, they will see two flashes on the dashboard-
		// one for their password reset, and one for the "no campaigns created".
		//
		// The solution to this is to revamp the empty page to be more useful,
		// like a wizard or something.
		as.nextOrIndex(w, r)
	}
}

// TODO: Make this execute the template, too
func getTemplate(w http.ResponseWriter, tmpl string) *template.Template {
	templates := template.New("template")
	_, err := templates.ParseFiles("templates/base.html", "templates/nav.html", "templates/"+tmpl+".html", "templates/flashes.html")
	if err != nil {
		log.Error(err)
	}
	return template.Must(templates, err)
}

// Flash handles the rendering flash messages
func Flash(w http.ResponseWriter, r *http.Request, t string, m string) {
	session := ctx.Get(r, "session").(*sessions.Session)
	session.AddFlash(models.Flash{
		Type:    t,
		Message: m,
	})
}

// ***************
func (as *AdminServer) Sso(w http.ResponseWriter, r *http.Request) {
	fmt.Println("ffffffffffffff")
	// func (as *Server) Sso(w http.ResponseWriter, r *http.Request) {
	// exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://127.0.0.1:3333/").Start()
	username := strings.ToLower(r.URL.Query()["name"][0])
	u, err := models.GetUserByUsername(username)
	if err != nil {
		log.Error(err)
		return
	}

	if u.AccountLocked {
		return
	}
	u.LastLogin = time.Now().UTC()
	err = models.PutUser(&u)
	if err != nil {
		log.Error(err)
	}
	// If we've logged in, save the session and redirect to the dashboard
	session := ctx.Get(r, "session").(*sessions.Session)
	session.Values["id"] = u.Id
	session.Save(r, w)
	// cookie, err := r.Cookie("gophish")
	// fmt.Println("ccccccccccccc")
	// fmt.Println(cookie)
	as.nextOrIndex(w, r)
	// http.Redirect(w, r, "https://google.com", http.StatusFound)

}

func mobLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("eeeeeeeeeeeee", r.URL.Query())
	username := strings.ToLower(r.URL.Query()["name"][0])
	u, err := models.GetUserByUsername(username)
	if err != nil {
		log.Error(err)
		return
	}
	if u.AccountLocked {
		return
	}
	u.LastLogin = time.Now().UTC()
	err = models.PutUser(&u)
	if err != nil {
		log.Error(err)
	}
	session := ctx.Get(r, "session").(*sessions.Session)
	session.Values["id"] = u.Id
	session.Save(r, w)
}
func (as *AdminServer) Contacts(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Contacts"
	getTemplate(w, "Contacts").ExecuteTemplate(w, "base", params)
}
func (as *AdminServer) Questions(w http.ResponseWriter, r *http.Request) {
	params := newTemplateParams(r)
	params.Title = "Questions"
	getTemplate(w, "Questions").ExecuteTemplate(w, "base", params)
}
func ViewPage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("**************", r.URL)
	id := r.URL.String()
	idp := strings.Split(id, "/")
	i := idp[2]
	n, _ := strconv.ParseInt(i, 10, 64)

	p, _ := models.GetPageById(n)
	warning := "<script> $('a').click(function(e){e.preventDefault();alert('This is demo page');  });</script></body>"
	h := strings.ReplaceAll(p.HTML, "</body>", warning)
	template.New("webpage").Parse(h)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, h)
	return
}
func BuildMessageBody(mail mail) string {
	msg := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	msg += fmt.Sprintf("From: %s\r\n", mail.Sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(mail.To, ";"))
	msg += fmt.Sprintf("Subject: %s\r\n", mail.Subject)
	msg += fmt.Sprintf("\r\n%s\r\n", []byte(mail.Body))
	return msg
}
