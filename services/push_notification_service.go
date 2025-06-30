package services

import (
	"context"
	"encoding/json"
	"fmt"

	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
	"github.com/pricetra/api/database/jet/postgres/public/model"
	"github.com/pricetra/api/database/jet/postgres/public/table"
)

func (s Service) CreatePushNotificationEntry(ctx context.Context, req any, res expo.PushResponse) (pn model.PushNotification, err error) {
	req_json, err := json.Marshal(req)
	if err != nil {
		return model.PushNotification{}, fmt.Errorf("could not marshal req data: %s", err.Error())
	}
	res_json, err := json.Marshal(res)
	if err != nil {
		return model.PushNotification{}, fmt.Errorf("could not marshal res data: %s", err.Error())
	}

	request := string(req_json)
	response := string(res_json)
	qb := table.PushNotification.INSERT(
		table.PushNotification.Request,
		table.PushNotification.Response,
	).MODEL(model.PushNotification{
		Request: &request,
		Response: &response,
	}).RETURNING(table.PushNotification.AllColumns)
	if err = qb.QueryContext(ctx, s.DB, &pn); err != nil {
		return model.PushNotification{}, err
	}
	return pn, nil
}
