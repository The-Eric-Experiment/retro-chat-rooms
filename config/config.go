package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type ConfigChatRoom struct {
	ID                   string `yaml:"id"`
	Name                 string `yaml:"name"`
	Description          string `yaml:"description"`
	Color                string `yaml:"color"`
	DiscordChannel       string `yaml:"discord-channel"`
	ChatRoomIntroMessage string `yaml:"chat-room-intro-message"`
}

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
	DiscordWebhookId     string              `yaml:"discord-webhook-id"`
	DiscordWebhookToken  string              `yaml:"discord-webhook-token"`
	OwnerChatUser        OwnerChatUserConfig `yaml:"owner-chat-user"`
	Rooms                []ConfigChatRoom    `yaml:"rooms"`
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

var Current = LoadConfig()
