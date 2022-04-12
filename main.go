package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"gopkg.in/yaml.v2"
)

type OwnerChatUserConfig struct {
	DiscordId string `yaml:"discord_id"`
	Id        string `yaml:"id"`
	Nickname  string `yaml:"nickname"`
	Color     string `yaml:"color"`
	Password  string `yaml:"password"`
}

type Config struct {
	SiteName             string              `yaml:"site-name"`
	ChatRoomHeaderLogo   string              `yaml:"chat-room-header-logo"`
	ChatRoomHeaderHeight string              `yaml:"chat-room-header-height"`
	DiscordBotToken      string              `yaml:"discord-bot-token"`
	OwnerChatUser        OwnerChatUserConfig `yaml:"owner-chat-user"`
	Rooms                []*Room             `yaml:"rooms"`
}

func LoadConfig() Config {
	yamlFile, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	var c Config
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

var session = Session{}

var config = LoadConfig()

var ownerRoomUser *RoomUser

func initializeOnwner(config OwnerChatUserConfig) {
	if config.DiscordId == "" {
		return
	}

	ownerRoomUser = &RoomUser{
		ID:                 config.Id,
		LastActivity:       time.Now().UTC(),
		Nickname:           config.Nickname,
		Color:              config.Color,
		LastPing:           time.Now().UTC(),
		LastUserListUpdate: time.Now().UTC(),
		SessionIdent:       "",
		IsDiscordUser:      config.DiscordId != "",
		IsOwner:            true,
	}
}

var discord = DiscordBot{}

var rooms = config.Rooms

func getAllTemplates() map[string]string {
	templates := make(map[string]string)
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ext := filepath.Ext(path)

			if ext == ".html" {
				name := filepath.Base(path)
				b, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				templates[name] = string(b)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return templates
}

func LoadTemplates(router *gin.Engine) {
	funcMap := template.FuncMap{
		"isUserNotNil":     isUserNotNil,
		"formatMessage":    formatMessage,
		"formatTime":       formatTime,
		"formatNickname":   formatNickname,
		"isMessageVisible": isMessageVisible,
		"isMessageToSelf":  isMessageToSelf,
		"countUsers":       countUsers,
		"hasStrings":       hasStrings,
	}

	templates := getAllTemplates()

	t := template.New("").Delims("{{", "}}").Funcs(funcMap)

	for filename, content := range templates {
		defineExpr, err := regexp.Compile("\\{\\{[\\s]*define[\\s]+\"[^\"]+\"[\\s]*}}")
		if err != nil {
			panic(err)
		}
		tagStartExpr, err := regexp.Compile("(>(?:[\\s\n\r])+)")
		if err != nil {
			panic(err)
		}

		tagEndExpr, err := regexp.Compile("((?:[\\s\n\r])+<)")
		if err != nil {
			panic(err)
		}
		spaceExpr, err := regexp.Compile("[\\s\r\n]+")
		if err != nil {
			panic(err)
		}
		contained := "{{ define \"" + filename + "\" }}" + content + "{{end}}"
		if defineExpr.MatchString(content) {
			contained = content
		}

		final := spaceExpr.ReplaceAllString(
			tagEndExpr.ReplaceAllString(
				tagStartExpr.ReplaceAllString(contained, ">"), "<",
			),
			" ",
		)

		t = template.Must(t.Parse(final))
	}

	router.SetHTMLTemplate(t)
}

func onDiscordConnected(user *discordgo.User) {
	fmt.Println(user)
}

func receiveDiscordMessage(m *discordgo.MessageCreate) {
	content := m.Content

	room, found := lo.Find(rooms, func(room *Room) bool { return room.DiscordChannel == m.ChannelID })

	if !found {
		return
	}

	user := room.GetUser(m.Author.ID)
	if user == nil {
		user = room.RegisterDiscordUser(m.Author)
	}
	now := time.Now().UTC()

	messageMentionExpr := regexp.MustCompile(`^\s*@([^:]+):\s*(.+)`)

	match := messageMentionExpr.FindStringSubmatch(m.Content)

	var to *RoomUser

	if len(match) > 1 {
		content = match[2]
		to = room.GetUserByNickname(match[1])
	}

	room.SendMessage(&RoomMessage{
		Time:            now,
		Message:         content,
		From:            user,
		SpeechMode:      MODE_SAY_TO,
		To:              to,
		IsSystemMessage: false,
		FromDiscord:     true,
	})

	room.Update()
}

func main() {
	LoadProfanityFilters()
	session.Initialize()
	initializeOnwner(config.OwnerChatUser)

	for _, room := range rooms {
		room.Initialize()
	}

	// Background Tasks
	go checkUserStatus()
	go sessionCleanup()

	router := gin.Default()

	router.Use(static.Serve("/public", static.LocalFile("./public", true)))

	// router.SetFuncMap()

	LoadTemplates(router)

	router.GET("/", GetIndex)
	router.GET("/room/:id", GetRoom)
	router.POST("/room/:id", PostRoom)
	router.GET("chat-login/:id", GetChatLogin)
	router.GET("/chat-header/:id", GetChatHeader)
	router.GET("/chat-thread/:id/:userId", GetChatThread)
	router.POST("/post-message", PostMessage)
	router.GET("/chat-updater/:id/:userId", GetChatUpdater)
	router.GET("/chat-talk/:id/:userId", GetChatTalk)
	router.GET("/chat-users/:id/:userId", GetChatUsers)
	router.GET("/chaptcha", GetChaptcha)
	router.POST("/logout", PostLogout)

	for _, room := range rooms {
		room.chatEventAwaiter.Initialize()
	}

	discord.Connect(onDiscordConnected)
	discord.OnReceiveMessage(receiveDiscordMessage)

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
	discord.Close()
}
