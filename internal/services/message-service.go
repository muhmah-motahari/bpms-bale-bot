package service

import (
	"bbb/configs"
	"bbb/internal/dto"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type (
	MessageService interface {
		SendStringMessage(chatID int64, txt string, replyID int)
	}

	messageService struct {
		env configs.Env
	}
)

func NewMessageService(env configs.Env) MessageService {
	return &messageService{
		env: env,
	}
}

func (m *messageService) SendStringMessage(chatID int64, txt string, replyID int) {
	ms := dto.SendMessage{
		Chat_id:             chatID,
		Text:                txt,
		Reply_to_message_id: replyID,
	}

	body, _ := json.Marshal(ms)

	url := fmt.Sprintf("%vsendMessage", m.env.BaleAPIAddress)
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(body))
}
