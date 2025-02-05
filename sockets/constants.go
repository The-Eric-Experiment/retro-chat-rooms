package sockets

const (
	SERVER_ERROR                     = 0
	SERVER_COLOR_LIST                = 1
	SERVER_ROOM_LIST                 = 2
	SERVER_USER_REGISTRATION_SUCCESS = 3
	SERVER_USER_LIST_ADD             = 4
	SERVER_USER_LIST_REMOVE          = 5
	SERVER_USER_LIST_UPDATE          = 6
	SERVER_MESSAGE_SENT              = 7
	SERVER_USER_KICKED               = 8
	SERVER_TIME                      = 9

	CLIENT_REGISTER_USER      = 100
	CLIENT_SEND_MESSAGE       = 101
	CLIENT_COLOR_LIST_REQUEST = 105
	CLIENT_ROOM_LIST_REQUEST  = 106
	CLIENT_PING               = 110
)

const SERVER_TIME_MIN = 5
