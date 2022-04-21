package routes

import (
	"log"
	"net/url"

	"github.com/google/uuid"
)

func BustCache(input string) string {
	// Use url.Parse() to parse a string into a *url.URL type. If your URL is
	// already a url.URL type you can skip this step.
	urlA, err := url.Parse(input)
	if err != nil {
		log.Fatal(err)
	}

	values := urlA.Query()

	values.Set("b", uuid.NewString())

	urlA.RawQuery = values.Encode()

	return urlA.String()
}

func UrlRoom(id string) string {
	return BustCache("/room/" + id)
}

func UrlJoin(id string) string {
	return BustCache("/join/" + id)
}

func UrlLogout() string {
	return BustCache("/logout")
}

func UrlCaptcha() string {
	return BustCache("/chaptcha")
}

func UrlChatHeader(id string) string {
	return BustCache("/chat-header/" + id)
}

func UrlChatThread(id string) string {
	return BustCache("/chat-thread/" + id)
}

func UrlChatUpdater(id string) string {
	return BustCache("/chat-updater/" + id)
}

func UrlChatTalk(id string, to string) string {
	urlA, err := url.Parse("/chat-talk/" + id)
	if err != nil {
		log.Fatal(err)
	}

	if to != "" {

		values := urlA.Query()

		values.Set("to", to)

		urlA.RawQuery = values.Encode()
	}

	return BustCache(urlA.String())
}

func UrlChatUsers(id string) string {
	return BustCache("/chat-users/" + id)
}
