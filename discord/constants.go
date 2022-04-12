package discord

import chatroom "retro-chat-rooms/chatroom"

var modeTransform = map[string]string{
	chatroom.MODE_SAY_TO:     "says to",
	chatroom.MODE_SCREAM_AT:  "SCREAMS AT",
	chatroom.MODE_WHISPER_TO: "whispers to",
}
