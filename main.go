package main

import (
	"fmt"
	"retro-chat-rooms/chatroom"
	"retro-chat-rooms/discord"
	"retro-chat-rooms/profanity"
	"retro-chat-rooms/routes"
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
	profanity.LoadProfanityFilters()
	// chatroom.InitializeOwner(config.Current.OwnerChatUser)

	// Background Tasks
	go tasks.CheckUserStatus(chatroom.CHAT_ROOMS)
	go tasks.ObserveMessagesForDiscord()

	router := gin.Default()

	store := cookie.NewStore([]byte("secret1"))
	router.Use(sessions.Sessions("chatsession", store))

	// router.Use(func(c *gin.Context) {
	// 	// Except for these
	// 	imgRegexp := regexp.MustCompile(`\.(gif|jpeg|jpg|png)$`)
	// 	if !imgRegexp.MatchString(c.Request.URL.Path) {
	// 		c.Header("Pragma", "no-cache")
	// 		c.Header("Cache-Control", "no-cache")
	// 	}
	// 	c.Next()
	// })

	router.Use(static.Serve("/public", static.LocalFile("./public", true)))

	templates.LoadTemplates(router)

	// Chat login
	router.GET("/", routes.GetIndex)
	router.GET("/join/:id", routes.GetJoin)
	router.POST("/join/:id", routeWithSession(routes.PostJoin))
	router.GET("/chaptcha", routeWithSession(routes.GetChaptcha))
	router.POST("/logout", routes.PostLogout)

	// Main chat Screen
	router.GET("/room/:id", routes.GetRoom)
	router.GET("/chat-header/:id", routes.GetChatHeader)
	router.GET("/chat-thread/:id", routes.GetChatThread)
	router.GET("/chat-updater/:id", routeWithSession(routes.GetChatUpdater))
	router.GET("/chat-talk/:id", routes.GetChatTalk)
	router.POST("/chat-talk/:id", routeWithSession(routes.PostChatTalk))
	router.GET("/chat-users/:id", routes.GetChatUsers)

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
