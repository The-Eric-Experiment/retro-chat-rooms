package main

import (
	"image/color"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/steambap/captcha"
)

const (
	MODE_SAY_TO     = "says-to"
	MODE_SCREAM_AT  = "screams-at"
	MODE_WHISPER_TO = "whispers-to"

	USER_COLOR_BLACK  = "#000000"
	USER_COLOR_BEIGE  = "#B5A642"
	USER_COLOR_BROWN  = "#5c3317"
	USER_COLOR_PINK   = "#FF00FF"
	USER_COLOR_PURPLE = "#800080"
	USER_COLOR_NAVY   = "#004080"
	USER_COLOR_TEAL   = "#008080"
	USER_COLOR_ORANGE = "#FF7F00"
	USER_COLOR_RED    = "#FF0000"
	USER_COLOR_BLUE   = "#0000FF"
)

var speechModes = []SpeechMode{
	{Value: MODE_SAY_TO, Label: "Say to"},
	{Value: MODE_SCREAM_AT, Label: "Scream at"},
	{Value: MODE_WHISPER_TO, Label: "Whisper to"},
}

var colors = []string{
	USER_COLOR_BLACK,
	USER_COLOR_BEIGE,
	USER_COLOR_BROWN,
	USER_COLOR_PINK,
	USER_COLOR_PURPLE,
	USER_COLOR_NAVY,
	USER_COLOR_TEAL,
	USER_COLOR_ORANGE,
	USER_COLOR_RED,
	USER_COLOR_BLUE,
}

func getIP(c *gin.Context) string {
	headers := [4]string{
		"HTTP_CF_CONNECTING_IP", "HTTP_X_REAL_IP", "HTTP_X_FORWARDED_FOR", "REMOTE_ADDR",
	}

	for _, headerKey := range headers {
		ip := c.Request.Header.Get(headerKey)
		if ip != "" {
			return ip
		}
	}

	return ""
}

func getSessionUserIdent(ctx *gin.Context) string {
	uagent := ctx.Request.Header.Get("User-Agent")
	ip := getIP(ctx)
	return uuid.NewSHA1(uuid.NameSpaceURL, []byte(ip+uagent)).String()
}

func checkUserStatus() {
	for {
		for _, room := range rooms {
			// Check if there are expired users
			for _, user := range room.Users {
				if time.Now().UTC().Sub(user.LastPing).Seconds() > USER_LOGOUT_TIMEOUT {
					room.DeregisterUser(user)
				}
			}
		}
		time.Sleep(20000 * time.Millisecond)
	}
}

func GetIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{})
}

func GetRoom(c *gin.Context) {
	id := c.Param("id")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-login.html", gin.H{
		"Colors": colors,
		"Color":  room.Color,
		"ID":     room.ID,
		"Name":   room.Name,
	})
}

func PostRoom(c *gin.Context) {
	id := c.Param("id")
	nickname := c.PostForm("ni")
	color := c.PostForm("co")
	captcha := c.PostForm("chap")

	userIdent := getSessionUserIdent(c)
	userSess := sessions[userIdent]
	sessionCaptcha := userSess["captcha"].(string)

	if sessionCaptcha != captcha {
		c.Status(http.StatusBadRequest)
		return
	}

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	user, err := room.RegisterUser(nickname, color)

	if err != nil {
		c.HTML(http.StatusOK, "chat-login.html", gin.H{
			"Colors": colors,
			"ID":     room.ID,
			"Name":   room.Name,
		})
		return
	}

	c.HTML(http.StatusOK, "main.html", gin.H{
		"UserID":   user.ID,
		"Name":     room.Name,
		"Color":    room.Color,
		"ID":       room.ID,
		"Nickname": user.Nickname,
	})
}

func GetChatLogin(c *gin.Context) {
	id := c.Param("id")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-login.html", gin.H{
		"Color":  room.Color,
		"Name":   room.Name,
		"Colors": colors,
		"ID":     room.ID,
	})
}

func GetChatHeader(c *gin.Context) {
	id := c.Param("id")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-header.html", gin.H{
		"Name":  room.Name,
		"Color": room.Color,
	})
}

func GetChatThread(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-thread.html", gin.H{
		"ID":       room.ID,
		"Name":     room.Name,
		"Color":    room.Color,
		"UserID":   userId,
		"Messages": room.Messages,
	})
}

func PostMessage(c *gin.Context) {
	id := c.PostForm("id")
	message := c.PostForm("message")
	userId := c.PostForm("userId")
	to := c.PostForm("to")
	mode := c.PostForm("speechMode")
	private := c.PostForm("private")

	if len(strings.TrimSpace(message)) == 0 {
		c.Redirect(302, "/chat-updater/"+id+"/"+userId)
		return
	}

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	user, ok := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	userTo, _ := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == to })

	room.SendMessage(&RoomMessage{
		Time:            time.Now().UTC(),
		Message:         message,
		From:            user,
		To:              userTo,
		IsSystemMessage: false,
		Privately:       private == "on",
		SpeechMode:      mode,
	})

	c.Redirect(302, "/chat-updater/"+room.ID+"/"+user.ID)
}

func GetChatUpdater(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	user, ok := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })

	if !ok {
		c.HTML(http.StatusOK, "chat-updater.html", gin.H{
			"UserGone": true,
			"ID":       room.ID,
		})
		return
	}

	UpdateMessageRate(room, user)

	_, hasMessages := lo.Find(
		room.Messages,
		func(m *RoomMessage) bool { return m.Time.Sub(user.LastPing).Seconds() > 0 },
	)

	userListUpdated := room.LastUserListUpdate.Sub(user.LastUserListUpdate).Seconds() > 0

	user.LastPing = time.Now().UTC()

	c.HTML(http.StatusOK, "chat-updater.html", gin.H{
		"ID":              room.ID,
		"UserID":          user.ID,
		"MessageRate":     user.UpdateMessageRate,
		"HasMessages":     hasMessages,
		"UserListUpdated": userListUpdated,
		"Color":           room.Color,
	})
}

func GetChatTalk(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")
	toUserId := c.Query("to")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	user, ok := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	var to *RoomUser = nil

	toUser, ok := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == toUserId })

	if ok {
		to = toUser
	}

	c.HTML(http.StatusOK, "chat-talk.html", gin.H{
		"ID":          room.ID,
		"UserId":      user.ID,
		"Nickname":    user.Nickname,
		"To":          to,
		"SpeechModes": speechModes,
	})
}

func GetChatUsers(c *gin.Context) {
	id := c.Param("id")
	userId := c.Param("userId")

	room, ok := lo.Find(rooms, func(r *Room) bool { return r.ID == id })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	user, ok := lo.Find(room.Users, func(r *RoomUser) bool { return r.ID == userId })

	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "chat-users.html", gin.H{
		"ID":          room.ID,
		"UserID":      user.ID,
		"Users":       room.Users,
		"UsersOnline": len(room.Users),
		"RoomTime":    time.Now().UTC(),
	})
}

func GetChaptcha(ctx *gin.Context) {
	ident := getSessionUserIdent(ctx)
	data, err := captcha.NewMathExpr(222, 122, func(options *captcha.Options) {
		options.FontScale = 1
		options.BackgroundColor = color.White
		options.CurveNumber = 20
		options.TextLength = 6
		options.FontDPI = 50
		options.Noise = 10
	})

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	sessions[ident] = UserSession{
		"captcha": data.Text,
	}

	data.WriteImage(ctx.Writer)
}
