package dto

type ChatDMPayload struct {
	RecipientUsername string `json:"recipient_username"`
	Content           string `json:"content"`
}
