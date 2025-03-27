package routes

import (
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/config"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func GetAdminLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "admin-login.html", gin.H{
		"Rooms": chat.GetAllRooms(),
	})
}

func PostAdminLogin(c *gin.Context, session sessions.Session) {
	nick := c.PostForm("u")
	pass := c.PostForm("p")
	roomId := c.PostForm("r")
	captcha := c.PostForm("mess")

	room, found := chat.GetSingleRoom(roomId)

	if !found {
		c.Redirect(http.StatusFound, BustCache("/"))
		return
	}

	isOwnerVariation := chat.IsNickVariation(config.Current.OwnerChatUser.Nickname, nick)

	hasher := sha1.New()
	hasher.Write([]byte(pass))
	hash := hex.EncodeToString(hasher.Sum(nil))

	ownerPassword := strings.ToLower(config.Current.OwnerChatUser.Password)

	sessionCaptcha := session.Get("chaptcha")

	captchaInvalid := sessionCaptcha == nil || sessionCaptcha.(string) != captcha

	if !isOwnerVariation || ownerPassword != hash || captchaInvalid {
		c.HTML(http.StatusForbidden, "admin-login.html", gin.H{
			"Rooms": chat.GetAllRooms(),
		})
		return
	}

	registeredUserId := chat.RegisterAdmin(roomId, userAgentToClientInfo(c.GetHeader("User-Agent")))

	userId := session.Get("userId")
	combinedId := chat.GetCombinedId(roomId, userId.(string))

	if userId == nil || combinedId != registeredUserId {
		session.Set("userId", config.Current.OwnerChatUser.Id)
		session.Save()
	}

	session.Set("supportsChatEventAwaiter", supportsChatEventAwaiter(c))
	session.Save()

	c.Redirect(http.StatusFound, UrlRoom(room.ID))
}
