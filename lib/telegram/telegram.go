package telegram

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
)

type Telegram interface {
	ParseIncomingRequest(reader io.Reader) (*Update, error)
	ReplyMessage(chatId int64, text string) error
	ReplyImage(chatId int64, name, caption string, image []byte) error
	GetWebhook() (string, error)
	SetWebhook(webhook string) error
}

type telegramImpl struct {
	token string
}

func NewTelegram(token string) Telegram {
	return &telegramImpl{
		token: token,
	}
}

func (self *telegramImpl) ParseIncomingRequest(reader io.Reader) (*Update, error) {
	var msgUpdate Update

	err := json.NewDecoder(reader).Decode(&msgUpdate)
	if err != nil {
		return nil, err
	}

	if msgUpdate.UpdateID == 0 {
		return nil, errors.New("invalid update request message")
	}

	return &msgUpdate, nil
}

func (self *telegramImpl) ReplyMessage(chatId int64, text string) error {
	replyObj := map[string]string{
		"chat_id":    strconv.FormatInt(chatId, 10),
		"text":       text,
		"parse_mode": "MarkdownV2",
	}
	replyMsg, err := json.Marshal(replyObj)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", self.token),
		"application/json",
		bytes.NewBuffer(replyMsg),
	)
	if err != nil {
		return err
	}
	defer func(body io.ReadCloser) {
		if err := body.Close(); err != nil {
			log.Println("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		errMsg := &APIResponse{}
		err := json.NewDecoder(resp.Body).Decode(&errMsg)

		if err != nil {
			return fmt.Errorf("Error parsing response: %v", err)
		}

		return fmt.Errorf("Status %q: %s", resp.Status, errMsg.Description)
	}
	return nil
}

func (self *telegramImpl) ReplyImage(
	chatId int64,
	name, caption string,
	image []byte,
) error {
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	err := w.WriteField("chat_id", strconv.FormatInt(chatId, 10))
	if err != nil {
		return err
	}

    w.WriteField("chat_id", fmt.Sprintf("%d", chatId))
    w.WriteField("caption", caption)

	part, err := w.CreateFormFile("photo", name)
	if err != nil {
		return err
	}

	part.Write(image)
	err = w.Close()
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf(
            "https://api.telegram.org/bot%s/sendPhoto", 
            self.token,
        ),
		w.FormDataContentType(),
		body,
	)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := &APIResponse{}
		err := json.NewDecoder(resp.Body).Decode(&errMsg)

		if err != nil {
			return fmt.Errorf("Error parsing response: %v", err)
		}

		return fmt.Errorf("Status %q: %s", resp.Status, errMsg.Description)
	}
	return nil
}
