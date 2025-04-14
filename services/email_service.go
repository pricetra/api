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
	"github.com/sendgrid/rest"
)

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
) (response *rest.Response, err error) {
	req, err := s.NewEmailPostRequest(ctx, "/email-verification", map[string]string{
		"recipientEmail": user.Email,
		"name": user.Name,
		"code": email_verification.Code,
	})
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if !(res.StatusCode >= 200 && res.StatusCode < 300) {
		return nil, fmt.Errorf(res.Status)
	}

	json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
