package chat

import (
	"fmt"
	"regexp"
	"strings"
)

func ExtractClientInfo(client string) ClientInfo {
	info := ClientInfo{}

	var re = regexp.MustCompile(`(?mi)(?<label>[^\s:]{1,5}):\((?<value>[^\)]+)\)`)

	labelIdx := re.SubexpIndex("label")
	valueIdx := re.SubexpIndex("value")

	for _, match := range re.FindAllStringSubmatch(client, -1) {
		label := match[labelIdx]
		value := match[valueIdx]

		switch label {
		case "os":
			info.OS = value
		case "plat":
			info.Plat = value
		case "env":
			info.Env = value
		case "v":
			info.Version = value
		}

	}

	return info
}

func FormatClientInfo(info ClientInfo) string {
	var result strings.Builder

	if info.Plat != "" {
		result.WriteString(fmt.Sprintf("plat:(%s) ", info.Plat))
	}

	if info.OS != "" {
		result.WriteString(fmt.Sprintf("os:(%s) ", info.OS))
	}

	if info.Env != "" {
		result.WriteString(fmt.Sprintf("env:(%s) ", info.Env))
	}

	if info.Version != "" {
		result.WriteString(fmt.Sprintf("v:(%s)", info.Version))
	}

	// Trim trailing space if there is one
	return strings.TrimSpace(result.String())
}

func ClientInfoToMsgSource(info ClientInfo) string {
	switch info.Plat {
	case CLIENT_PLATFORM_DESKTOP:
		if info.Env == "16-bit" {
			return MSG_SOURCE_WINDOWS_16
		}
		return ""
	case CLIENT_PLATFORM_WEB:
		return MSG_SOURCE_WEB
	case CLIENT_PLATFORM_DISCORD:
		return MSG_SOURCE_DISCORD
	}

	return ""
}
