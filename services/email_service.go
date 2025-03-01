package services

import (
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (service Service) SendTemplateEmail(
	to *mail.Email, 
	subject string, 
	template_id string, 
	template_data map[string]any,
) (*rest.Response, error) {
	from := mail.NewEmail("Stadio", "no-reply@thestadio.com")
	email := mail.NewV3MailInit(from, subject, to)
	email.SetTemplateID(template_id)

	personalization := mail.NewPersonalization()
	personalization.AddTos(to)
	for key, val := range template_data {
		personalization.SetDynamicTemplateData(key, val)
	}
	email.AddPersonalizations(personalization)
	return service.Sendgrid.Send(email)
}

func (service Service) SendEmailVerification(user gmodel.User, email_verification model.EmailVerification) (*rest.Response, error) {
	to := mail.NewEmail(user.Name, user.Email)
	data := map[string]any{
		"user": map[string]any{
			"id": user.ID,
			"name": user.Name,
			"email": user.Email,
		},
		"code": email_verification.Code,
	}
	return service.SendTemplateEmail(
		to,
		"Stadio: Email verification code", 
		service.Tokens.SendGrid.Templates.EmailVerification, 
		data,
	)
}
