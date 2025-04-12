package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/sendgrid/rest"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func template_params(html string, data map[string]string) string {
	result := html
	for key, val := range data {
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), val)
	}
	return result
}

func (s Service) NewEmailSender() *mail.Email {
	return mail.NewEmail("Pricetra", "no-reply@pricetra.com")
}

func (s Service) SendPlainTextEmail(
	ctx context.Context,
	to *mail.Email,
	subject string,
	content string,
) (*rest.Response, error) {
	email := mail.NewSingleEmailPlainText(s.NewEmailSender(), subject, to, content)
	return s.Sendgrid.SendWithContext(ctx, email)
}

func (s Service) SendHtmlEmail(
	ctx context.Context,
	to *mail.Email,
	subject string,
	content string,
	html string,
	data map[string]string,
) (*rest.Response, error) {
	html = template_params(html, data)
	email := mail.NewSingleEmail(s.NewEmailSender(), subject, to, content, html)
	return s.Sendgrid.SendWithContext(ctx, email)
}

// This function is buggy due to Sendgrid not supporting templates 
// with subjects and personalization data.
// See https://github.com/sendgrid/sendgrid-go/issues/341
func (s Service) SendTemplateEmail(
	ctx context.Context,
	to *mail.Email,
	subject string,
	template_id string,
	template_data map[string]any,
) (*rest.Response, error) {
	email := mail.NewV3MailInit(s.NewEmailSender(), subject, to)
	email.SetTemplateID(template_id)

	personalization := mail.NewPersonalization()
	personalization.AddTos(to)
	for key, val := range template_data {
		personalization.SetDynamicTemplateData(key, val)
	}
	email.AddPersonalizations(personalization)
	return s.Sendgrid.SendWithContext(ctx, email)
}

func (s Service) SendEmailVerification(
	ctx context.Context,
	user gmodel.User,
	email_verification model.EmailVerification,
) (*rest.Response, error) {
	to := mail.NewEmail(user.Name, user.Email)
	html := `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd"><html dir="ltr" lang="en"><head><link rel="preload" as="image" href="https://res.cloudinary.com/pricetra-cdn/image/upload/v1743137791/logotype_black_color_cckqyj.png"><link rel="preload" as="image" href="https://res.cloudinary.com/pricetra-cdn/image/upload/v1744403096/logotype_white_color_cckqyj.png"><meta content="text/html; charset=UTF-8" http-equiv="Content-Type"><meta name="x-apple-disable-message-reformatting"><!--$--><style>@media (prefers-color-scheme:light){.light{display:block!important}.dark{display:none!important}}</style></head><div style="display:none;overflow:hidden;line-height:1px;opacity:0;max-height:0;max-width:0">Your verification code is {{code}}<div></div></div><body style="background-color:#fff;font-family:-apple-system,BlinkMacSystemFont,&quot"><table align="center" width="100%" border="0" cellpadding="0" cellspacing="0" role="presentation" style="margin-top:2.5rem;margin-bottom:2.5rem;margin-left:auto;margin-right:auto;padding-top:2.5rem;padding-bottom:2.5rem;padding-left:1.25rem;padding-right:1.25rem;background-color:#fff;max-width:450px;border-radius:.5rem"><tbody><tr style="width:100%"><td><table align="center" width="100%" border="0" cellpadding="0" cellspacing="0" role="presentation" style="padding-left:3rem;padding-right:3rem"><tbody><tr><td><table align="center" width="100%" border="0" cellpadding="0" cellspacing="0" role="presentation" style="display:flex;align-items:center;justify-content:center;text-align:center"><tbody><tr><td><img class="light" alt="Pricetra" src="https://res.cloudinary.com/pricetra-cdn/image/upload/v1743137791/logotype_black_color_cckqyj.png" style="display:block;outline:0;border:none;text-decoration:none" width="180"><img class="dark" alt="Pricetra" src="https://res.cloudinary.com/pricetra-cdn/image/upload/v1744403096/logotype_white_color_cckqyj.png" style="display:none;outline:0;border:none;text-decoration:none" width="180"></td></tr></tbody></table><hr style="border-color:#e5e7eb;margin-top:2rem;margin-bottom:.5rem;width:100%;border:none;border-top:1px solid #eaeaea"><table align="center" width="100%" border="0" cellpadding="0" cellspacing="0" role="presentation"><tbody><tr><td><p style="font-size:1.5rem;line-height:2rem;font-weight:700;margin-bottom:16px;margin-top:16px">Hi, <!-- -->{{name}}</p><p style="color:#374151;font-size:14px;line-height:24px;margin-bottom:16px;margin-top:16px">We&#x27;re almost done, you just need to verify your email address using the code below to activate your Pricetra account.</p><p style="font-size:1.5rem;line-height:2rem;margin-top:1.25rem;text-align:center;text-indent:20px;letter-spacing:20px;background-color:#f3f4f6;padding:1.25rem;border-radius:.5rem;font-weight:700;margin-bottom:16px">{{code}}</p></td></tr></tbody></table></td></tr></tbody></table></td></tr></tbody></table><!--/$--></body></html>`
	return s.SendHtmlEmail(
		ctx,
		to,
		"Email Verification Code", 
		fmt.Sprintf("Your email verification code is: %s", email_verification.Code), 
		html,
		map[string]string{
			"name": user.Name,
			"code": email_verification.Code,
		},
	)
}
