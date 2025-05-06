package dto

type (
	CallbackQuery struct {
		ID      string  `json:"id"`
		From    User    `json:"from"`
		Message Message `json:"message"`
		Data    string  `json:"data"`
	}
)
