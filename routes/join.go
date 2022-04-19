package routes

import (
	"net/http"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/session"

	"github.com/gin-gonic/gin"
)

func getJoinData(c *gin.Context, room *chatroom.Room, errors []string) *gin.H {
	if room == nil {
		id := c.Param("id")
		room = chatroom.FindRoomByID(id)
	}

	if room == nil {
		return nil
	}

	return &gin.H{
		"Errors":      errors,
		"Colors":      chatroom.NICKNAME_COLORS,
		"Color":       room.Color,
		"TextColor":   room.TextColor,
		"Description": room.Description,
		"ID":          room.ID,
		"Name":        room.Name,
	}
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

	if len(errors) == 0 {
		_, err := room.RegisterUser(nickname, color, userIdent)
		if err != nil && err.Error() == "user exists" {
			errors = append(errors, "Someone is already using this Nickname, try a different one.")
		} else if err != nil {
			errors = append(errors, "Couldn't register user, try again.")
		}
	}

	// credentials := strings.Split(nickname, "###")
	// if len(credentials) > 1 {
	// 	loginNick := credentials[0]
	// 	loginPassword := credentials[1]
	// 	hasher := sha1.New()
	// 	hasher.Write([]byte(loginPassword))
	// 	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))

	// 	if sha != config.Current.OwnerChatUser.Password || config.Current.OwnerChatUser.Nickname == loginNick {

	// 	}
	// }

	if len(errors) > 0 {
		data := getJoinData(c, room, errors)
		c.HTML(http.StatusOK, "join.html", data)
		return
	}

	c.Redirect(http.StatusFound, "/room/"+room.ID)
}
