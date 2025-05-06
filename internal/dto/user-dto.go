package dto

const (
	USER_ROLE_NORMAL = "normal"
	USER_ROLE_ADMIN  = "admin"
)

type (
	User struct {
		ID            int64
		Is_bot        bool
		First_name    string
		Last_name     string
		Username      string
		Language_code string
	}

	GetMe struct {
		Response
		User User `json:"result"`
	}
)
