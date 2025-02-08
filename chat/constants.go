package chat

const (
	MAX_MESSAGES                      = 200
	USER_STALE_TIMEOUT                = 120
	USER_MAX_MESSAGE_RATE_SEC         = 60
	USER_MIN_MESSAGE_RATE_SEC         = 2
	USER_SCREAM_TIMEOUT_MIN   float64 = 2
	UPDATER_WAIT_TIMEOUT_MS           = 30000
	MAX_ROOM_MESSAGE_HISTORY          = 10

	MODE_SAY_TO     = "says-to"
	MODE_SCREAM_AT  = "screams-at"
	MODE_WHISPER_TO = "whispers-to"

	USER_COLOR_BLACK  = "#000000"
	USER_COLOR_BEIGE  = "#878700"
	USER_COLOR_BROWN  = "#875f00"
	USER_COLOR_PINK   = "#FF00AF"
	USER_COLOR_PURPLE = "#800080"
	USER_COLOR_NAVY   = "#004080"
	USER_COLOR_TEAL   = "#008080"
	USER_COLOR_ORANGE = "#FFAF00"
	USER_COLOR_RED    = "#FF0000"
	USER_COLOR_BLUE   = "#0000FF"
)

var SPEECH_MODES = []SpeechMode{
	{Value: MODE_SAY_TO, Label: "Say to"},
	{Value: MODE_SCREAM_AT, Label: "Scream at"},
	{Value: MODE_WHISPER_TO, Label: "Whisper to"},
}

var NICKNAME_COLORS = []NicknameColor{
	{Color: USER_COLOR_BLACK, Name: "Black"},
	{Color: USER_COLOR_BEIGE, Name: "Beige"},
	{Color: USER_COLOR_BROWN, Name: "Brown"},
	{Color: USER_COLOR_PINK, Name: "Pink"},
	{Color: USER_COLOR_PURPLE, Name: "Purple"},
	{Color: USER_COLOR_NAVY, Name: "Navy"},
	{Color: USER_COLOR_TEAL, Name: "Teal"},
	{Color: USER_COLOR_ORANGE, Name: "Orange"},
	{Color: USER_COLOR_RED, Name: "Red"},
	{Color: USER_COLOR_BLUE, Name: "Blue"},
}
