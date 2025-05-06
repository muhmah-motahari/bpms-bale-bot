package controller

import (
	"bbb/configs"
	"bbb/internal/dto"
	"bbb/service"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type (
	GetUpdatesController interface {
		GetUpdate(w http.ResponseWriter, r *http.Request)
	}

	getUpdatesController struct {
		messageService service.MessageService
		env            configs.Env
		groupService   service.GroupService
		userService    service.UserService
	}
)

func NewGetUpdatesController(ms service.MessageService, env configs.Env, gs service.GroupService, us service.UserService) GetUpdatesController {
	return &getUpdatesController{
		messageService: ms,
		env:            env,
		groupService:   gs,
		userService:    us,
	}
}

func (c *getUpdatesController) GetUpdate(w http.ResponseWriter, r *http.Request) {
	var req dto.Update

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(`{message: %v}`, err)))
		return
	}

	if len(req.Message.New_chat_members) > 0 {
		// Add to group
		if req.Message.New_chat_members[0].ID == c.env.BotID {
			c.groupService.AddToNewGroup(req.Message)
			c.messageService.SendStringMessage(req.Message.Chat.ID, "به به به به، اقا دمتون گرم", req.Message.Message_id)
		}
		// else -> Add new memeber to Database
	}
}
