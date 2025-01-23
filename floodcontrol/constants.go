package floodcontrol

const (
	// how long the user has to wait to send
	// a message after a flood in minutes
	MESSAGE_FLOOD_COOLDOWN_SEC float64 = 120
	// it's considered a flood if the number
	// of messages reached MESSAGE_FLOOD_MESSAGES_COUNT
	// within:
	MESSAGE_FLOOD_RANGE_SEC float64 = 5
	// it's considered a flood if the number
	// of messages is this within x seconds
	MESSAGE_FLOOD_MESSAGES_COUNT = 10
	// this is the number of floods the user
	// can do before they get blocked for a
	// longer period.
	MESSAGE_FLOOD_MAXIMUM_FLOODS = 4
	// this is how long the user gets locked
	// from sending messages if they flooded
	// more than the number above.
	MESSAGE_FLOOD_IP_BAN_MIN = 30
)
