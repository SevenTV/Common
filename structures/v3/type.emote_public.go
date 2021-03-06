package structures

import (
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PublicEmote struct {
	ID        primitive.ObjectID `json:"id"`
	OwnerID   primitive.ObjectID `json:"owner_id"`
	Name      string             `json:"name"`
	Flags     EmoteFlag          `json:"flags"`
	Tags      []string           `json:"tags"`
	Images    []string           `json:"images"`
	CreatedAt time.Time          `json:"created_at"`
}

func (e *Emote) ToPublic(cdnBase string) PublicEmote {
	version := e.GetLatestVersion(true)
	images := []string{}

	for _, file := range version.ImageFiles {
		if version.Animated && file.FrameCount == 1 {
			continue
		}

		images = append(images, fmt.Sprintf("//%s/%s", cdnBase, file.Key))
	}

	return PublicEmote{
		ID:        e.ID,
		OwnerID:   e.OwnerID,
		Name:      e.Name,
		Flags:     e.Flags,
		Tags:      e.Tags,
		Images:    images,
		CreatedAt: version.CreatedAt,
	}
}

type PublicEmoteFile struct {
	Name       string            `json:"name"`
	Format     PublicEmoteFormat `json:"format"`
	Width      int32             `json:"width"`
	Height     int32             `json:"height"`
	URL        string            `json:"url"`
	FrameCount int32             `json:"frame_count"`
	Size       int64             `json:"size"`
}

func (ef EmoteFile) ToPublic() PublicEmoteFile {
	var format PublicEmoteFormat
	if s := strings.Split(ef.ContentType, "image/"); len(s) == 2 {
		format = PublicEmoteFormat(strings.ToUpper(s[1]))
	}

	return PublicEmoteFile{
		Name:       ef.Name,
		Format:     format,
		Width:      ef.Width,
		Height:     ef.Height,
		FrameCount: ef.FrameCount,
		Size:       ef.Size,
	}
}

type PublicEmoteFormat string

const (
	PublicEmoteFormatAVIF PublicEmoteFormat = "AVIF"
	PublicEmoteFormatWEBP PublicEmoteFormat = "WEBP"
	PublicEmoteFormatGIF  PublicEmoteFormat = "GIF"
	PublicEmoteFormatPNG  PublicEmoteFormat = "PNG"
)
