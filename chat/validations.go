package chat

import (
	"math"
	"regexp"
	"retro-chat-rooms/config"
	"retro-chat-rooms/floodcontrol"
	"retro-chat-rooms/profanity"
	"strconv"
	"strings"
	"time"

	"github.com/samber/lo"
)

type IUserState interface {
	GetUserID() string
	GetCoolDownMessageSent() bool
	SetCoolDownMessageSent(v bool)
	GetLastScream() time.Time
	SetLastScream(t time.Time)
	GetUserIP() string
}

func checkForMessageAbuse(user ChatUser) bool {
	// Ignore checks for Discord users or admins
	if user.IsDiscordUser() || user.IsAdmin {
		return false
	}

	// Retrieve messages sent by the user
	messages, found := GetMessagesByUser(user.ID)
	if !found {
		return false
	}

	// Filter out system messages and only include user messages
	messages = lo.Filter(messages, func(m *ChatMessage, _ int) bool {
		return !m.IsSystemMessage && m.From == user.ID
	})

	// Check if we have enough messages to evaluate flood control
	if len(messages) < floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT {
		return false
	}

	// Find the most recent message
	lastMessage := messages[len(messages)-1]

	// Only consider messages within the last `MESSAGE_FLOOD_RANGE_SEC`
	thresholdTime := lastMessage.Time.Add(-time.Duration(floodcontrol.MESSAGE_FLOOD_RANGE_SEC) * time.Second)
	relevantMessages := lo.Filter(messages, func(m *ChatMessage, _ int) bool {
		return m.Time.After(thresholdTime)
	})

	// If there are fewer than the required count, no flood
	if len(relevantMessages) < floodcontrol.MESSAGE_FLOOD_MESSAGES_COUNT {
		return false
	}

	// Ensure the cooldown period has passed since the last message in the flood
	firstFloodMessage := relevantMessages[0] // The first message in the flood
	cooldownEnd := firstFloodMessage.Time.Add(time.Duration(floodcontrol.MESSAGE_FLOOD_COOLDOWN_SEC) * time.Second)

	// If the cooldown has expired and no further flood occurred, it's not abuse
	if time.Now().After(cooldownEnd) {
		return false
	}

	// Otherwise, the user is still in a restricted state
	return true
}

// substringPercentage finds the start and end index of a substring within a larger string
// and calculates the percentage of characters the substring represents in the larger string.
func substringPercentage(str, substr string) (percentage float64) {
	startIndex := strings.Index(str, substr)
	if startIndex == -1 {
		return 0
	}

	percentage = (float64(len(substr)) / float64(len(str))) * 100
	return percentage
}

// LevenshteinDistance calculates the Levenshtein distance between two strings
func LevenshteinDistance(a, b string) int {
	la := len(a)
	lb := len(b)

	// Initialize the matrix
	d := make([][]int, la+1)
	for i := range d {
		d[i] = make([]int, lb+1)
	}

	// Set up the base cases
	for i := 0; i <= la; i++ {
		d[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		d[0][j] = j
	}

	// Fill in the matrix
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			substitutionCost := 1
			if a[i-1] == b[j-1] {
				substitutionCost = 0
			}
			d[i][j] = int(math.Min(
				math.Min(float64(d[i-1][j]+1), float64(d[i][j-1]+1)),
				float64(d[i-1][j-1]+substitutionCost),
			))
		}
	}

	return d[la][lb]
}

func IsNickVariation(input string, against string) bool {
	m1 := regexp.MustCompile(`[\s_-]+`)
	inputStr := m1.ReplaceAllString(strings.ToLower(strings.TrimSpace(input)), "")
	againstStr := m1.ReplaceAllString(strings.ToLower(strings.TrimSpace(against)), "")
	percentA := substringPercentage(inputStr, againstStr)
	percentB := substringPercentage(againstStr, inputStr)
	distance := LevenshteinDistance(strings.ToLower(inputStr), strings.ToLower(againstStr))
	distancePercent := float64(distance) / float64(len(againstStr)) // lower is closer
	return percentA >= 70.0 || percentB >= 70.0 || distancePercent <= 0.25
}

func ValidateMessage(userState IUserState, inputMsg ChatMessage) (ChatMessage, bool) {
	now := time.Now().UTC()

	room, found := GetSingleRoom(inputMsg.RoomID)

	if !found {
		return ChatMessage{}, false
	}

	user := inputMsg.GetFrom()

	if user == nil {
		return ChatMessage{}, false
	}

	if len(strings.TrimSpace(inputMsg.Message)) == 0 {
		return ChatMessage{}, false
	}

	userIp := userState.GetUserIP()

	if checkForMessageAbuse(*user) && !floodcontrol.IsCooldownPeriod(userIp) {
		floodcontrol.SetFlooded(userIp)
	}

	if floodcontrol.IsIPBanned(userIp) {
		return ChatMessage{
			RoomID:               room.ID,
			Time:                 time.Now().UTC(),
			Message:              "{nickname} was kicked for flooding the channel too many times.",
			IsSystemMessage:      true,
			SystemMessageSubject: user,
			SpeechMode:           MODE_SAY_TO,
		}, true
	}

	coolDownMessageSent := userState.GetCoolDownMessageSent()

	// Maybe with some refactor we can use the checkForMessageAbuse directly here?
	if floodcontrol.IsCooldownPeriod(userIp) && coolDownMessageSent {
		return ChatMessage{}, false
	}

	if floodcontrol.IsCooldownPeriod(userIp) && !coolDownMessageSent {
		userState.SetCoolDownMessageSent(true)

		return ChatMessage{
			RoomID:               room.ID,
			Time:                 now,
			To:                   user.ID,
			IsSystemMessage:      true,
			Message:              "Hey {nickname}, chill out, you'll be able to send messages again in " + strconv.FormatInt(int64(floodcontrol.MESSAGE_FLOOD_COOLDOWN_SEC)/60, 10) + " minutes.",
			Privately:            true,
			SystemMessageSubject: user,
			SpeechMode:           MODE_SAY_TO,
			InvolvedUsers:        []ChatUser{*user},
		}, true
	}

	userState.SetCoolDownMessageSent(false)

	// Check if there's slurs

	if profanity.HasBlockedWords(inputMsg.Message) {
		return ChatMessage{
			RoomID:               room.ID,
			Time:                 now,
			To:                   user.ID,
			IsSystemMessage:      true,
			Message:              "Come on {nickname}! Let's be nice! This is a place for having fun!",
			Privately:            true,
			SystemMessageSubject: user,
			SpeechMode:           MODE_SAY_TO,
			InvolvedUsers:        []ChatUser{*user},
		}, true
	}

	// Check if user has screamed recently

	lastScream := userState.GetLastScream()

	if now.Sub(lastScream).Minutes() <= USER_SCREAM_TIMEOUT_MIN && inputMsg.SpeechMode == MODE_SCREAM_AT {
		return ChatMessage{
			RoomID:               room.ID,
			Time:                 now,
			To:                   user.ID,
			IsSystemMessage:      true,
			SystemMessageSubject: user,
			Message:              "Hi {nickname}, you're only allowed to scream once every " + strconv.FormatInt(int64(USER_SCREAM_TIMEOUT_MIN), 10) + " minutes.",
			Privately:            true,
			SpeechMode:           MODE_SAY_TO,
			InvolvedUsers:        []ChatUser{*user},
		}, true

	}

	if inputMsg.SpeechMode == MODE_SCREAM_AT {
		userState.SetLastScream(now)
	}

	involvedUsers := []ChatUser{*user}
	toUser, foundToUser := GetUser(inputMsg.To)

	if foundToUser {
		involvedUsers = append(involvedUsers, toUser)
	}

	message := profanity.ReplaceSensoredProfanity(inputMsg.Message)

	return ChatMessage{
		RoomID:          room.ID,
		Time:            now,
		Message:         message,
		From:            user.ID,
		To:              inputMsg.To,
		IsSystemMessage: false,
		Privately:       inputMsg.Privately,
		SpeechMode:      inputMsg.SpeechMode,
		InvolvedUsers:   involvedUsers,
	}, true
}

func ValidateUser(userState IUserState, user ChatUser, errors *[]string) {
	if floodcontrol.IsIPBanned(userState.GetUserIP()) {
		*errors = append(*errors, "You have been temporarily kicked out for flooding, try again later.")
	}

	if user.Nickname == "" {
		*errors = append(*errors, "You must provide a Nickname.")
	}

	validNickname, err := regexp.MatchString(`^[a-zA-Z0-9\s_-]+$`, user.Nickname)

	if !validNickname || err != nil {
		*errors = append(*errors, "Only alpha-numeric characters, spaces, underscores and dashes are allowed in nicknames.")
	}

	if profanity.IsProfaneNickname(user.Nickname) {
		*errors = append(*errors, "This nickname is not allowed.")
	}

	isOwnerVariation := IsNickVariation(user.Nickname, config.Current.OwnerChatUser.Nickname)
	_, hasUserNickname := GetUserByNickname(user.Nickname)

	if isOwnerVariation || hasUserNickname {
		*errors = append(*errors, "Someone is already using this Nickname, try a different one.")
	}

	_, hasUser := GetUser(user.ID)

	if hasUser {
		*errors = append(*errors, "User already logged in.")
	}
}
