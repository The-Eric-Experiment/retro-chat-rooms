package discord

import "retro-chat-rooms/chat"

var modeTransform = map[string]string{
	chat.MODE_SAY_TO:     "says to",
	chat.MODE_SCREAM_AT:  "SCREAMS AT",
	chat.MODE_WHISPER_TO: "whispers to",
}
