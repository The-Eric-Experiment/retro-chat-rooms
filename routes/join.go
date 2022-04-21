package routes

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/http"
	"regexp"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/session"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ua-parser/uap-go/uaparser"
)

func supportsChatEventAwaiter(c *gin.Context) bool {
	parser, err := uaparser.New("./ua_regexes.yaml")
	if err != nil {
		log.Fatal(err)
	}

	client := parser.Parse(c.GetHeader("User-Agent"))
	family := strings.ToLower(client.UserAgent.Family)
	major, err := strconv.ParseInt(client.UserAgent.Major, 10, 0)

	if err != nil {
		major = 99
	}

	if family == "opera" && major <= 4 {
		return false
	}

	if family == "ie" && major <= 3 {
		return false
	}

	return true
}

func getJoinData(c *gin.Context, room *chatroom.Room, errors []string) *gin.H {
	if room == nil {
		id := c.Param("id")
		room = chatroom.FindRoomByID(id)
	}

	if room == nil {
		return nil
	}

	return &gin.H{
		"CaptchaBuster": uuid.NewString(),
		"Errors":        errors,
		"Colors":        chatroom.NICKNAME_COLORS,
		"Color":         room.Color,
		"TextColor":     room.TextColor,
		"Description":   room.Description,
		"ID":            room.ID,
		"Name":          room.Name,
	}
}

func isNickVariation(input string, against string) bool {
	m1 := regexp.MustCompile(`[\s_-]+`)
	inputStr := m1.ReplaceAllString(strings.ToLower(strings.TrimSpace(input)), "")
	againstStr := m1.ReplaceAllString(strings.ToLower(strings.TrimSpace(against)), "")

	return inputStr == againstStr
}

func GetJoin(c *gin.Context) {
	data := getJoinData(c, nil, make([]string, 0))
	if data == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "join.html", data)
}

func PostJoin(c *gin.Context) {
	id := c.Param("id")
	ni := c.PostForm("ni")
	color := c.PostForm("co")
	captcha := c.PostForm("chap")

	room := chatroom.FindRoomByID(id)

	credentials := strings.Split(ni, "###")

	nickname := credentials[0]

	if room == nil {
		c.Redirect(http.StatusFound, BustCache("/"))
		return
	}

	userIdent := session.GetSessionUserIdent(c)
	session.RegisterUserIP(c)
	sessionCaptcha, foundCaptcha := session.GetSessionValue(userIdent, "captcha")

	errors := make([]string, 0)

	if session.IsIPBanned(userIdent) {
		errors = append(errors, "You have been temporarily kicked out for flooding, try again later.")
	}

	if nickname == "" {
		errors = append(errors, "No nickname was entered in the form.")
	}

	if profanity.IsProfaneNickname(nickname) {
		errors = append(errors, "This nickname is not allowed.")
	}

	if !foundCaptcha || sessionCaptcha.(string) != captcha {
		errors = append(errors, "The entered captcha is invalid.")
	}

	isOwnerVariation := isNickVariation(config.Current.OwnerChatUser.Nickname, nickname)

	// YUCK THESE IFS
	if len(errors) == 0 {
		if len(credentials) > 1 {
			loginPassword := credentials[1]
			hasher := sha1.New()
			hasher.Write([]byte(loginPassword))
			hash := hex.EncodeToString(hasher.Sum(nil))

			ownerPassword := strings.ToLower(config.Current.OwnerChatUser.Password)

			if hash == ownerPassword && isOwnerVariation {
				chatroom.LoginOwner(config.Current.OwnerChatUser, userIdent)
			} else {
				errors = append(errors, "Entered credentials are invalid!")
			}
		} else if isOwnerVariation {
			errors = append(errors, "Someone is already using this Nickname, try a different one.")
		} else {
			_, err := room.RegisterUser(nickname, color, userIdent)
			if err != nil && err.Error() == "user exists" {
				errors = append(errors, "Someone is already using this Nickname, try a different one.")
			} else if err != nil {
				errors = append(errors, "Couldn't register user, try again.")
			}
		}
	}

	if len(errors) > 0 {
		data := getJoinData(c, room, errors)
		c.HTML(http.StatusOK, "join.html", data)
		return
	}

	session.SetSessionValue(userIdent, "supportsChatEventAwaiter", supportsChatEventAwaiter(c))

	c.Redirect(http.StatusFound, UrlRoom(room.ID))
}
