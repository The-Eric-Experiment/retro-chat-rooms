package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/session"

	"github.com/gin-gonic/gin"
)

func GetRoom(c *gin.Context) {
	id := c.Param("id")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Status(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "join.html", gin.H{
		"Errors":      make([]string, 0),
		"Colors":      chatroom.NICKNAME_COLORS,
		"Color":       room.Color,
		"TextColor":   room.TextColor,
		"Description": room.Description,
		"ID":          room.ID,
		"Name":        room.Name,
	})
}

func PostRoom(c *gin.Context) {
	id := c.Param("id")
	nickname := c.PostForm("ni")
	color := c.PostForm("co")
	captcha := c.PostForm("chap")

	room := chatroom.FindRoomByID(id)

	if room == nil {
		c.Redirect(http.StatusTemporaryRedirect, "/")
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

	var user *chatroom.RoomUser
	if len(errors) == 0 {
		var err error
		user, err = room.RegisterUser(nickname, color, userIdent)
		if err != nil && err.Error() == "user exists" {
			errors = append(errors, "Someone is already using this Nickname, try a different one.")
		} else if err != nil {
			errors = append(errors, "Couldn't register user, try again.")
		}
	}

	if len(errors) > 0 {
		c.HTML(http.StatusOK, "join.html", gin.H{
			"Errors":    errors,
			"Colors":    chatroom.NICKNAME_COLORS,
			"Color":     room.Color,
			"TextColor": room.TextColor,
			"ID":        room.ID,
			"Name":      room.Name,
		})
		return
	}

	c.HTML(http.StatusOK, "main.html", gin.H{
		"UserID":       user.ID,
		"Name":         room.Name,
		"Color":        room.Color,
		"ID":           room.ID,
		"Nickname":     user.Nickname,
		"HeaderHeight": config.Current.ChatRoomHeaderHeight,
	})
}
