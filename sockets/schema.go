package sockets

type RoomListRequest struct {
}

type RoomList struct {
	RoomIDs   []string `fieldOrder:"0"`
	RoomNames []string `fieldOrder:"1"`
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
	Client   string `fieldOrder:"3"`
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

type ServerUserKicked struct {
	Reason string `fieldOrder:"0"`
}

type ServerUserListUpdate struct {
	RoomID  string `fieldOrder:"0"`
	TotalID int32  `fieldOrder:"1"`
}

type ServerMessageSent struct {
	RoomID               string `fieldOrder:"0"`
	From                 string `fieldOrder:"1"`
	To                   string `fieldOrder:"2"`
	Privately            string `fieldOrder:"3"`
	SpeechMode           string `fieldOrder:"4"`
	Time                 string `fieldOrder:"5"`
	IsSystemMessage      string `fieldOrder:"6"`
	SystemMessageSubject string `fieldOrder:"7"`
	IsHistory            string `fieldOrder:"8"`
	Message              string `fieldOrder:"9"`
}
