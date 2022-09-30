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
	Images    []PublicImage      `json:"images"`
	CreatedAt time.Time          `json:"created_at"`
}

func (e *Emote) ToPublic(cdnBase string) PublicEmote {
	version := e.GetLatestVersion(true)
	images := []PublicImage{}

	for _, file := range version.ImageFiles {
		if version.Animated && file.FrameCount == 1 {
			continue
		}

		f := file.ToPublic()
		f.URL = fmt.Sprintf("//%s/%s", cdnBase, file.Key)

		images = append(images, f)
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

type PublicImage struct {
	Name       string            `json:"name"`
	Format     PublicEmoteFormat `json:"format"`
	Width      int32             `json:"width"`
	Height     int32             `json:"height"`
	URL        string            `json:"url"`
	FrameCount int32             `json:"frame_count"`
	Size       int64             `json:"size"`
}

func (ef ImageFile) ToPublic() PublicImage {
	var format PublicEmoteFormat
	if s := strings.Split(ef.ContentType, "image/"); len(s) == 2 {
		format = PublicEmoteFormat(strings.ToUpper(s[1]))
	}

	return PublicImage{
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
