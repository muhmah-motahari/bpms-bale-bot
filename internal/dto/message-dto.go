package dto

type (
	Message struct {
		Message_id              int      `json:"message_id"`
		From                    User     `json:"from"`
		Date                    int      `json:"date"`
		Chat                    Chat     `json:"chat"`
		Forward_from            User     `json:"forward_from"`
		Forward_from_chat       Chat     `json:"forward_from_chat"`
		Forward_from_message_id int      `json:"forward_from_message_id"`
		Forward_date            int      `json:"forward_date"`
		Reply_to_message        *Message `json:"reply_to_message"`
		Edite_date              int      `json:"edite_date"`
		Text                    string   `json:"text"`
		Caption                 string   `json:"caption"`
		New_chat_members        []User   `json:"new_chat_members"`
		Left_chat_member        User     `json:"left_chat_member"`
	}

	SendMessage struct {
		Chat_id             int64  `json:"chat_id"`
		Text                string `json:"text"`
		Reply_to_message_id int    `json:"reply_to_message_id"`
	}

	SendMessageResponse struct {
		Response
		SendMessage
	}
)
