package utils

import (
	"fmt"

	"github.com/SevenTV/Common/configure"
)

func GetCdnURL(emoteID string, size int8) string {
	return fmt.Sprintf("%v/emote/%v/%dx", configure.Config.GetString("cdn_url"), emoteID, size)
}