package dto

type (
	Update struct {
		ID            int           `json:"update_id"`
		Message       Message       `json:"message"`
		EditedMessage Message       `json:"edited_message"`
		CallbackQuery CallbackQuery `json:"callback_query"`
	}

	UpdateResponse struct {
		OK     bool     `json:"ok"`
		Result []Update `json:"result"`
	}
)
