package routes

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetAdminLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-login.html", gin.H{
		"Rooms": chatroom.CHAT_ROOMS,
	})
}

func PostAdminLogin(c *gin.Context, session sessions.Session) {
	nick := c.PostForm("u")
	pass := c.PostForm("p")
	roomId := c.PostForm("r")
	captcha := c.PostForm("mess")

	room := chatroom.FindRoomByID(roomId)

	if room == nil {
		c.Redirect(http.StatusFound, BustCache("/"))
		return
	}

	isOwnerVariation := isNickVariation(config.Current.OwnerChatUser.Nickname, nick)

	hasher := sha1.New()
	hasher.Write([]byte(pass))
	hash := hex.EncodeToString(hasher.Sum(nil))

	ownerPassword := strings.ToLower(config.Current.OwnerChatUser.Password)

	sessionCaptcha := session.Get("chaptcha")

	captchaInvalid := sessionCaptcha == nil || sessionCaptcha.(string) != captcha

	if !isOwnerVariation || ownerPassword != hash || captchaInvalid {
		c.HTML(http.StatusForbidden, "admin-login.html", gin.H{
			"Rooms": chatroom.CHAT_ROOMS,
		})
		return
	}

	user := room.RegisterOwner(false)

	userId := session.Get("userId")

	if userId == nil || userId != user.ID {
		userId = uuid.NewString()
		session.Set("userId", user.ID)
		session.Save()
	}

	session.Set("supportsChatEventAwaiter", supportsChatEventAwaiter(c))
	session.Save()

	c.Redirect(http.StatusFound, UrlRoom(room.ID))
}
