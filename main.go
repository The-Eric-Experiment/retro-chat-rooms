package main

import (
	"fmt"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/config"
	"retro-chat-rooms/discord"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/routes"
	"retro-chat-rooms/tasks"
	"retro-chat-rooms/templates"

	nocache "github.com/alexander-melentyev/gin-nocache"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
)

func main() {
	profanity.LoadProfanityFilters()
	chatroom.InitializeOwner(config.Current.OwnerChatUser)

	// Background Tasks
	go tasks.CheckUserStatus(chatroom.CHAT_ROOMS)
	go tasks.SessionCleanup(chatroom.CHAT_ROOMS)
	go tasks.ObserveMessagesForDiscord()

	router := gin.Default()

	router.Use(nocache.NoCache())

	router.Use(static.Serve("/public", static.LocalFile("./public", true)))

	templates.LoadTemplates(router)

	// Chat login
	router.GET("/", routes.GetIndex)
	router.GET("/join/:id", routes.GetJoin)
	router.POST("/join/:id", routes.PostJoin)
	router.GET("/chaptcha", routes.GetChaptcha)
	router.POST("/logout", routes.PostLogout)

	// Main chat Screen
	router.GET("/room/:id", routes.GetRoom)
	router.GET("/chat-header/:id", routes.GetChatHeader)
	router.GET("/chat-thread/:id", routes.GetChatThread)
	router.GET("/chat-updater/:id", routes.GetChatUpdater)
	router.GET("/chat-talk/:id", routes.GetChatTalk)
	router.GET("/chat-users/:id", routes.GetChatUsers)
	router.POST("/post-message", routes.PostMessage)

	discord.Instance.Connect()
	discord.Instance.OnReceiveMessage(routes.ReceiveDiscordMessage)

	// go func() {
	// 	fmt.Println("Server is now running.  Press CTRL-C to exit.")
	// 	sc := make(chan os.Signal, 1)
	// 	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	// 	<-sc

	// 	fmt.Println("Disconnecting from Discord")
	// 	discord.Close()
	// }()

	router.Run()
	fmt.Println("closing...")
	discord.Instance.Close()
}
