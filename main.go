package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

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

var session = Session{}

var config = LoadConfig()

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

func main() {
	LoadProfanityFilters()
	session.Initialize()

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

	router.Run()
}
