package api

type ChatRoom struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Color  string `json:"color"`
	Online int    `json:"online"`
}
