package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/graph/gmodel"
	"github.com/pricetra/api/oapi"
)

func (s Service) NewEmailClient() (*oapi.ClientWithResponses, error) {
	return oapi.NewClientWithResponses(
		s.Tokens.EmailServer.Url, 
		oapi.WithRequestEditorFn(
			func(ctx context.Context, req *http.Request) error {
				req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Tokens.EmailServer.ApiKey))
				return nil
			},
		),
	)
}

func (s Service) NewEmailPostRequest(ctx context.Context, endpoint string, input map[string]string) (*http.Request, error)  {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	url := s.Tokens.EmailServer.Url + endpoint
	jsonInput, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonInput))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.Tokens.EmailServer.ApiKey))
	return req, nil
}

func (s Service) SendEmailVerification(
	ctx context.Context,
	user gmodel.User,
	email_verification model.EmailVerification,
) (*oapi.SendEmailVerificationCodeResponse, error) {
	client, err := s.NewEmailClient()
	if err != nil {
		return nil, err
	}
	res, err := client.SendEmailVerificationCodeWithResponse(ctx, oapi.EmailVerificationRequest{
		RecipientEmail: &user.Email,
		Name: &user.Name,
		Code: &email_verification.Code,
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}
