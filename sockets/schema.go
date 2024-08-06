package sockets

type RoomListRequest struct {
}

type RoomListStart struct {
}

type RoomListEnd struct {
}

type RoomListItem struct {
	RoomId string `fieldOrder:"0"`
	Name   string `fieldOrder:"1"`
}

type SendMessage struct {
	UserID     string `fieldOrder:"0"`
	To         string `fieldOrder:"1"`
	SpeechMode string `fieldOrder:"2"`
	Message    string `fieldOrder:"3"`
	Privately  string `fieldOrder:"4"`
	RoomID     string `fieldOrder:"5"`
}

type ColorListRequest struct {
}

type ColorList struct {
	Colors     []string `fieldOrder:"0"`
	ColorNames []string `fieldOrder:"1"`
}

type Ping struct {
	UserId string `fieldOrder:"0"`
}

type ServerError struct {
	Message string `fieldOrder:"0"`
}

type RegisterUser struct {
	Nickname string `fieldOrder:"0"`
	Color    string `fieldOrder:"1"`
	RoomID   string `fieldOrder:"2"`
}

type RegisterUserResponse struct {
	UserID   string `fieldOrder:"0"`
	Nickname string `fieldOrder:"1"`
	Color    string `fieldOrder:"2"`
	RoomID   string `fieldOrder:"3"`
}

type ServerUserListAdd struct {
	UserID   string `fieldOrder:"0"`
	Nickname string `fieldOrder:"1"`
	Color    string `fieldOrder:"2"`
	RoomID   string `fieldOrder:"3"`
}

type ServerUserListRemove struct {
	UserID string `fieldOrder:"0"`
	RoomID string `fieldOrder:"1"`
}

type ServerUserListUpdate struct {
	RoomID  string `fieldOrder:"0"`
	TotalID int32  `fieldOrder:"1"`
}

type ServerMessageSent struct {
	FromNickname string `fieldOrder:"0"`
	FromID       string `fieldOrder:"1"`
	ToNickname   string `fieldOrder:"2"`
	ToID         string `fieldOrder:"3"`
	Text         string `fieldOrder:"4"`
}
