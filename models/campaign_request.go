package models

type TestCampaign struct {
	Name       string   `json:"name"`
	Template   Template `json:"template"`
	TemplateId int64    `json:"template_id"`
	Url        string   `json:"url"`
	Page       Page     `json:"page"`
	PageId     int64    `json:"page_id"`
	SMTP       SMTP     `json:"smtp"`
	SmtpName   string   `json:"smtp_name"`
	FirstName  string   `json:"first_name"`
	LastName   string   `json:"last_name"`
	Email      string   `json:"email"`
	Position   string   `json:"position"`
}

type ResponseBody struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Errore  string `json:"errore"`
}
