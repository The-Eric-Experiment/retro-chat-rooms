package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"retro-chat-rooms/api"
	"retro-chat-rooms/chat"
	"retro-chat-rooms/discord"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/routes"
	"retro-chat-rooms/sockets"
	"retro-chat-rooms/tasks"
	"retro-chat-rooms/templates"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func routeWithSession(fn func(ctx *gin.Context, session sessions.Session)) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		session := sessions.Default(ctx)
		fn(ctx, session)
	}
}

func main() {
	gob.Register(map[string]string{})

	profanity.LoadProfanityFilters()

	chat.InitializeRooms()

	// Background Tasks
	go tasks.CheckUserStatus()
	tasks.ObserveMessagesToDiscord()
	discord.Instance.Connect()
	discord.Instance.OnReceiveMessage(tasks.OnReceiveDiscordMessage)

	router := gin.Default()

	store := cookie.NewStore([]byte("secret1"))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 365,
		HttpOnly: true,
		Secure:   false, // Change to true if using HTTPS
	})
	router.Use(sessions.Sessions("chatsession", store))

	router.Use(static.Serve("/public", static.LocalFile("./public", true)))
	router.Use(static.Serve("/assets", static.LocalFile("./assets", true)))

	templates.LoadTemplates(router)

	// Chat login
	router.GET("/", routes.GetIndex)
	router.GET("/join/:id", routeWithSession(routes.GetJoin))
	router.POST("/join/:id", routeWithSession(routes.PostJoin))
	router.POST("/logout", routeWithSession(routes.PostLogout))
	router.GET("/admin-login", routes.GetAdminLogin)
	router.POST("/admin-login", routeWithSession(routes.PostAdminLogin))

	// Main chat Screen
	router.GET("/room/:id", routeWithSession(routes.GetRoom))
	router.GET("/chat-header/:id", routes.GetChatHeader)
	router.GET("/chat-thread/:id", routeWithSession(routes.GetChatThread))
	router.GET("/chat-updater/:id", routeWithSession(routes.GetChatUpdater))
	router.GET("/chat-talk/:id", routeWithSession(routes.GetChatTalk))
	router.POST("/chat-talk/:id", routeWithSession(routes.PostChatTalk))
	router.GET("/chat-users/:id", routeWithSession(routes.GetChatUsers))

	// API
	group := router.Group("/api")
	{
		group.GET("/rooms", api.GetRooms)
	}

	go sockets.ServeSockets("8081")

	log.Panicln(router.Run())
	fmt.Println("closing...")
	discord.Instance.Close()
}
