package floodcontrol

const (
	// how long the user has to wait to send
	// a message after a flood in minutes
	MESSAGE_FLOOD_COOLDOWN_MIN float64 = 2
	// it's considered a flood if the number
	// of messages reached MESSAGE_FLOOD_MESSAGES_COUNT
	// within:
	MESSAGE_FLOOD_RANGE_SEC float64 = 4
	// it's considered a flood if the number
	// of messages is this within x seconds
	MESSAGE_FLOOD_MESSAGES_COUNT = 6
	// this is the number of floods the user
	// can do before they get blocked for a
	// longer period.
	MESSAGE_FLOOD_MAXIMUM_FLOODS = 5
	// this is how long the user gets locked
	// from sending messages if they flooded
	// more than the number above.
	MESSAGE_FLOOD_IP_BAN_MIN = 30
)
