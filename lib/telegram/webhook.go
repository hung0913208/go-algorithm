package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (self *telegramImpl) GetWebhook() (string, error) {
	apiMsg := &APIResponse{}
	webhookMsg := &WebhookInfo{}

	resp, err := http.Get(
		fmt.Sprintf("https://api.telegram.org/bot%s/getWebhookInfo", self.token),
	)
	if err != nil {
		return "", err
	}

	err = json.NewDecoder(resp.Body).Decode(&apiMsg)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", errors.New(apiMsg.Description)
	}

	err = json.Unmarshal(apiMsg.Result, &webhookMsg)
	if err != nil {
		return "", err
	}
	return webhookMsg.URL, nil
}

func (self *telegramImpl) SetWebhook(webhook string) error {
	errMsg := &APIResponse{}
	resp, err := http.Get(
		fmt.Sprintf(
			"https://api.telegram.org/bot%s/setWebhook?url=%s",
			self.token,
			webhook,
		),
	)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(errMsg.Description)
	}
	return nil
}
