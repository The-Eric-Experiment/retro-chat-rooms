package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

type Config struct {
	SiteName string  `yaml:site-name`
	Rooms    []*Room `yaml:"rooms"`
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

var config = LoadConfig()

var rooms = config.Rooms

type UserSession map[string]interface{}

var sessions = make(map[string]UserSession)

func getAllTemplates() []string {
	templates := make([]string, 0)
	err := filepath.Walk(".",
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			ext := filepath.Ext(path)

			if ext == ".html" {
				templates = append(templates, path)
			}

			return nil
		})
	if err != nil {
		log.Println(err)
	}

	return templates
}

func main() {
	go checkUserStatus()

	router := gin.Default()

	router.Use(static.Serve("/public", static.LocalFile("./public", true)))

	router.SetFuncMap(template.FuncMap{
		"isUserNotNil":     isUserNotNil,
		"formatMessage":    formatMessage,
		"formatTime":       formatTime,
		"formatNickname":   formatNickname,
		"isMessageVisible": isMessageVisible,
		"isMessageToSelf":  isMessageToSelf,
	})

	templates := getAllTemplates()

	router.LoadHTMLFiles(templates...)

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

	router.Run()
}
