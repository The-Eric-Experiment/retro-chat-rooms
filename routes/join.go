package routes

import (
	"log"
	"net/http"
	"retro-chat-rooms/chat"
	"strconv"
	"strings"

	"github.com/gin-contrib/sessions"
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

func getJoinData(room chat.ChatRoom, postUrl string, errors []string) *gin.H {
	return &gin.H{
		"CaptchaBuster": uuid.NewString(),
		"Errors":        errors,
		"Colors":        chat.NICKNAME_COLORS,
		"Color":         room.Color,
		"TextColor":     room.TextColor,
		"Description":   room.Description,
		"ID":            room.ID,
		"Name":          room.Name,
		"PostUrl":       postUrl,
	}
}

func GetJoin(c *gin.Context) {
	roomId := c.Param("id")
	room, found := chat.GetSingleRoom(roomId)
	if !found {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "join.html", getJoinData(room, UrlJoin(roomId), make([]string, 0)))
}

func validateAndJoin(c *gin.Context, session sessions.Session, room chat.ChatRoom, urlPost string) *gin.H {
	nickname := c.PostForm("ni")
	color := c.PostForm("co")
	captcha := c.PostForm("chap")

	userId := session.Get("userId")

	if userId == nil {
		userId = uuid.NewString()
		session.Set("userId", userId)
		session.Save()
	}

	sessionCaptcha := session.Get("chaptcha")

	errors := make([]string, 0)

	sessionUserState := NewSessionUserState(c, session)

	newUser := chat.ChatUser{
		RoomId:    room.ID,
		ID:        chat.GetCombinedId(room.ID, userId.(string)),
		Nickname:  nickname,
		Color:     color,
		DiscordId: "",
		IsAdmin:   false,
		IsWebUser: true,
	}

	if sessionCaptcha == nil || sessionCaptcha.(string) != captcha {
		errors = append(errors, "The entered captcha is invalid.")
	}

	chat.ValidateUser(&sessionUserState, newUser, &errors)

	// YUCK THESE IFS
	if len(errors) == 0 {
		_, err := chat.RegisterUser(newUser)
		if err != nil && err.Error() == "user exists" {
			errors = append(errors, "Someone is already using this Nickname, try a different one.")
		} else if err != nil {
			errors = append(errors, "Couldn't register user, try again.")
		}
	}

	if len(errors) > 0 {
		return getJoinData(room, urlPost, errors)
	}

	session.Set("supportsChatEventAwaiter", supportsChatEventAwaiter(c))
	session.Save()

	c.Redirect(http.StatusFound, UrlRoom(room.ID))
	return nil
}

func PostJoin(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")

	room, found := chat.GetSingleRoom(roomId)

	if !found {
		c.Redirect(http.StatusFound, BustCache("/"))
		return
	}

	failure := validateAndJoin(c, session, room, UrlJoin(room.ID))

	if failure == nil {
		return
	}

	c.HTML(http.StatusOK, "join.html", failure)
}
