package routes

import (
	"image/color"
	"image/gif"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/steambap/captcha"
)

func GetChaptcha(ctx *gin.Context, session sessions.Session) {
	data, err := captcha.NewMathExpr(222, 122, func(options *captcha.Options) {
		options.FontScale = 1
		options.BackgroundColor = color.White
		options.CurveNumber = 20
		options.TextLength = 6
		options.FontDPI = 50
		options.Noise = 10
	})

	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	session.Set("chaptcha", data.Text)
	session.Save()

	data.WriteGIF(ctx.Writer, &gif.Options{
		NumColors: 256,
	})
}
