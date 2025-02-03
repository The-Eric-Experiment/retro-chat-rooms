package routes

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"retro-chat-rooms/chat"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ua-parser/uap-go/uaparser"
)

// Field names to be randomized
var baseFieldNames = []string{"nickname", "color", "captcha", "email", "message", "phone", "address", "zip_code", "country", "city", "birthdate", "gender"}

func generateFieldNames() map[string]string {
	rand.Seed(time.Now().UnixNano())
	shuffled := make([]string, len(baseFieldNames))
	copy(shuffled, baseFieldNames)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	fieldMap := make(map[string]string)
	for i, key := range baseFieldNames {
		fieldMap[key] = shuffled[i]
	}
	return fieldMap
}

func getWestCoastDayOfMonth() int {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		log.Printf("Error loading timezone: %v", err)
		return time.Now().Day()
	}
	return time.Now().In(loc).Day()
}

func generateCaptcha() (int, int) {
	a := rand.Intn(8999) + 1000
	return a, a + getWestCoastDayOfMonth()
}

func parseUserAgent(userAgent string) (string, int64) {
	parser, err := uaparser.New("./ua_regexes.yaml")
	if err != nil {
		log.Printf("Error initializing UA parser: %v", err)
		return "", 0
	}

	client := parser.Parse(userAgent)
	family := strings.ToLower(client.UserAgent.Family)
	major, err := strconv.ParseInt(client.UserAgent.Major, 10, 0)
	if err != nil {
		major = 99
	}

	return family, major
}

func supportsChatEventAwaiter(c *gin.Context) bool {
	family, major := parseUserAgent(c.GetHeader("User-Agent"))

	if family == "opera" && major <= 4 {
		return false
	}

	if family == "ie" && major <= 3 {
		return false
	}

	return true
}

func getJoinData(session sessions.Session, room chat.ChatRoom, postUrl string, errors []string) *gin.H {
	fieldNames := generateFieldNames()
	session.Set("fieldNames", fieldNames)
	if err := session.Save(); err != nil {
		log.Println("Error saving session:", err)
	}
	a, sum := generateCaptcha()
	session.Set("captcha", sum)
	if err := session.Save(); err != nil {
		log.Println("Error saving session:", err)
	}

	return &gin.H{
		"Errors":      errors,
		"Colors":      chat.NICKNAME_COLORS,
		"Color":       room.Color,
		"TextColor":   room.TextColor,
		"Description": room.Description,
		"ID":          room.ID,
		"Name":        room.Name,
		"PostUrl":     postUrl,
		"FieldNames":  fieldNames,
		"CaptchaA":    a,
	}
}

func GetJoin(c *gin.Context, session sessions.Session) {
	roomId := c.Param("id")
	room, found := chat.GetSingleRoom(roomId)
	if !found {
		c.Status(http.StatusNotFound)
		return
	}
	c.HTML(http.StatusOK, "join.html", getJoinData(session, room, UrlJoin(roomId), make([]string, 0)))
}

func validateAndJoin(c *gin.Context, session sessions.Session, room chat.ChatRoom, urlPost string) *gin.H {
	fieldNamesInterface := session.Get("fieldNames")
	if fieldNamesInterface == nil {
		return getJoinData(session, room, urlPost, []string{"Session expired, please try again."})
	}

	fieldNames, ok := fieldNamesInterface.(map[string]string)
	if !ok {
		log.Println("Error: fieldNames is not of expected type")
		return getJoinData(session, room, urlPost, []string{"An error occurred, please try again."})
	}

	nickname := c.PostForm(fieldNames["nickname"])
	color := c.PostForm(fieldNames["color"])
	captchaInput := c.PostForm(fieldNames["captcha"])
	userId := session.Get("userId")
	if userId == nil {
		userId = uuid.NewString()
		session.Set("userId", userId)
		session.Save()
	}
	sessionCaptcha := session.Get("captcha")
	errors := make([]string, 0)

	if sessionCaptcha == nil || strconv.Itoa(sessionCaptcha.(int)) != captchaInput {
		errors = append(errors, "The entered captcha is invalid.")
	}

	sessionUserState := NewSessionUserState(c, session)

	browserFamily, browserMajor := parseUserAgent(c.GetHeader("User-Agent"))

	newUser := chat.ChatUser{
		RoomId:    room.ID,
		ID:        chat.GetCombinedId(room.ID, userId.(string)),
		Nickname:  nickname,
		Color:     color,
		DiscordId: "",
		IsAdmin:   false,
		IsWebUser: true,
		Client:    fmt.Sprintf("env:(Web) b:(%s) bv:(%d)", browserFamily, browserMajor),
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
		return getJoinData(session, room, urlPost, errors)
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
