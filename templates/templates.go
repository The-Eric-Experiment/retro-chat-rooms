package templates

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"retro-chat-rooms/routes"

	"github.com/gin-gonic/gin"
)

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
		"renderMessage":  RenderMessage,
		"renderUsername": RenderUsername,
		"formatTime":     formatTime,
		"countUsers":     countUsers,
		"hasStrings":     hasStrings,
		"bustCache":      routes.BustCache,
		"urlRoom":        routes.UrlRoom,
		"urlJoin":        routes.UrlJoin,
		"urlLogout":      routes.UrlLogout,
		"urlCaptcha":     routes.UrlCaptcha,
		"urlChatHeader":  routes.UrlChatHeader,
		"urlChatThread":  routes.UrlChatThread,
		"urlChatUpdater": routes.UrlChatUpdater,
		"urlChatTalk":    routes.UrlChatTalk,
		"urlChatUsers":   routes.UrlChatUsers,
	}

	templates := getAllTemplates()

	t := template.New("").Delims("{{", "}}").Funcs(funcMap)

	for filename, content := range templates {
		defineExpr := regexp.MustCompile("\\{\\{[\\s]*define[\\s]+\"[^\"]+\"[\\s]*}}")
		tagStartExpr := regexp.MustCompile("(>(?:[\\s\n\r])+)")
		tagEndExpr := regexp.MustCompile("((?:[\\s\n\r])+<)")
		spaceExpr := regexp.MustCompile("[\\s\r\n]+")
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
