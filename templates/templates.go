package templates

import (
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

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
