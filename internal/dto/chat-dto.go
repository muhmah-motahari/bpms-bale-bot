package dto

type (
	Chat struct {
		ID         int64  `json:"id"`
		Type       string `json:"type"`
		Title      string `json:"title"`
		Username   string `json:"username"`
		First_name string `json:"first_name"`
		Last_name  string `json:"last_name"`
	}

	GetChat struct {
		Chat_id string `json:"chat_id"`
	}
)
