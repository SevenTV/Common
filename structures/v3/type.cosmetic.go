package structures

import (
	"github.com/seventv/common/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CosmeticData interface {
	bson.Raw | CosmeticDataBadge | CosmeticDataPaint
}

type Cosmetic[D CosmeticData] struct {
	ID       primitive.ObjectID   `json:"id" bson:"_id"`
	Kind     CosmeticKind         `json:"kind" bson:"kind"`
	Priority int                  `json:"priority" bson:"priority"`
	Name     string               `json:"name" bson:"name"`
	UserIDs  []primitive.ObjectID `json:"users" bson:"user_ids"`
	Users    []User               `json:"user_objects" bson:"user_objects,skip,omitempty"`
	Data     D                    `json:"data" bson:"data"`

	// User Relationals
	Selected bool `json:"selected" bson:"selected,skip,omitempty"`
}

func (c Cosmetic[D]) ToRaw() Cosmetic[bson.Raw] {
	switch x := utils.ToAny(c.Data).(type) {
	case bson.Raw:
		return Cosmetic[bson.Raw]{
			ID:       c.ID,
			Kind:     c.Kind,
			Priority: c.Priority,
			Name:     c.Name,
			UserIDs:  c.UserIDs,
			Users:    c.Users,
			Data:     x,
			Selected: c.Selected,
		}
	}

	raw, _ := bson.Marshal(c.Data)
	return Cosmetic[bson.Raw]{
		ID:       c.ID,
		Kind:     c.Kind,
		Priority: c.Priority,
		Name:     c.Name,
		UserIDs:  c.UserIDs,
		Users:    c.Users,
		Data:     raw,
		Selected: c.Selected,
	}
}

func ConvertCosmetic[D CosmeticData](c Cosmetic[bson.Raw]) (Cosmetic[D], error) {
	var d D
	err := bson.Unmarshal(c.Data, &d)
	c2 := Cosmetic[D]{
		ID:       c.ID,
		Kind:     c.Kind,
		Priority: c.Priority,
		Name:     c.Name,
		UserIDs:  c.UserIDs,
		Users:    c.Users,
		Data:     d,
		Selected: c.Selected,
	}

	return c2, err
}

type CosmeticKind string

const (
	CosmeticKindBadge        CosmeticKind = "BADGE"
	CosmeticKindNametagPaint CosmeticKind = "PAINT"
	CosmeticKindAvatar       CosmeticKind = "AVATAR"
)

type CosmeticDataBadge struct {
	ID      primitive.ObjectID `json:"id" bson:"-"`
	Tag     string             `json:"tag" bson:"tag"`
	Tooltip string             `json:"tooltip" bson:"tooltip"`
	Misc    bool               `json:"misc,omitempty" bson:"misc"`
}

type CosmeticDataPaint struct {
	ID primitive.ObjectID `json:"id" bson:"-"`
	// The default color of the paint
	Color *utils.Color `json:"color" bson:"color"`
	// The canvas size for the paint
	CanvasSize [2]float32 `json:"canvas_size" bson:"canvas_size"`
	// The repeat mode of the canvas
	CanvasRepeat CosmeticPaintCanvasRepeat `json:"canvas_repeat" bson:"canvas_repeat"`
	// A list of gradients. There may be any amount, which can be stacked onto each other
	Gradients []CosmeticPaintGradient `json:"gradients" bson:"gradients"`
	// A list of drop shadows. There may be any amount, which can be stacked onto each other
	DropShadows []CosmeticPaintDropShadow `json:"drop_shadows,omitempty" bson:"drop_shadows,omitempty"`
	// A list of flairs
	Flairs []CosmeticPaintFlair `json:"flairs,omitempty" bson:"flairs,omitempty"`
	// Text stroke
	// Deprecated: The function used to generate the paint (i.e gradients or an image)
	Function CosmeticPaintGradientFunction `json:"function" bson:"function"`
	// Deprecated: Gradient stops, a list of positions and colors
	Stops []CosmeticPaintGradientStop `json:"stops" bson:"stops"`
	// Deprecated: Whether or not the gradient repeats outside its original area
	Repeat bool `json:"repeat" bson:"repeat"`
	// Deprecated: Gradient angle in degrees
	Angle int32 `json:"angle" bson:"angle"`
	// Deprecated: Shape of a radial gradient, when the paint is of RADIAL_GRADIENT type
	Shape string `json:"shape,omitempty" bson:"shape,omitempty"`
	// Deprecated: URL of an image, when the paint is of BACKGROUND_IMAGE type
	ImageURL string `json:"image_url,omitempty" bson:"image_url,omitempty"`
}

type CosmeticPaintGradient struct {
	// The function used to generate the paint (i.e gradients or an image)
	Function CosmeticPaintGradientFunction `json:"function" bson:"function"`
	// Gradient stops, a list of positions and colors
	Stops []CosmeticPaintGradientStop `json:"stops" bson:"stops"`
	// For a URL-based paint, the URL to an image
	ImageURL string `json:"image_url,omitempty" bson:"image_url,omitempty"`
	// For a radial gradient, the shape of the gradient
	Shape string `json:"shape,omitempty" bson:"shape,omitempty"`
	// The degree angle of the gradient (does not apply if function is URL)
	Angle int32 `json:"angle,omitempty" bson:"angle,omitempty"`
	// Whether or not the gradient repeats outside its original area
	Repeat bool `json:"repeat" bson:"repeat"`

	// Gradient position (X/Y % values)
	At [2]float64 `json:"at,omitempty" bson:"at,omitempty"`
}

type CosmeticPaintGradientFunction string

const (
	CosmeticPaintFunctionLinearGradient CosmeticPaintGradientFunction = "LINEAR_GRADIENT"
	CosmeticPaintFunctionRadialGradient CosmeticPaintGradientFunction = "RADIAL_GRADIENT"
	CosmeticPaintFunctionImageURL       CosmeticPaintGradientFunction = "URL"
)

type CosmeticPaintCanvasRepeat string

const (
	CosmeticPaintCanvasRepeatNone   CosmeticPaintCanvasRepeat = "no-repeat"
	CosmeticPaintCanvasRepeatX      CosmeticPaintCanvasRepeat = "repeat-x"
	CosmeticPaintCanvasRepeatY      CosmeticPaintCanvasRepeat = "repeat-y"
	CosmeticPaintCanvasRepeatRevert CosmeticPaintCanvasRepeat = "revert"
	CosmeticPaintCanvasRepeatRound  CosmeticPaintCanvasRepeat = "round"
	CosmeticPaintCanvasRepeatSpace  CosmeticPaintCanvasRepeat = "space"
)

type CosmeticPaintGradientStop struct {
	At    float64     `json:"at" bson:"at"`
	Color utils.Color `json:"color" bson:"color"`
	// the center position for the gradient. X/Y % values (for radial gradients only)
	CenterAt [2]float64 `json:"center_at,omitempty" bson:"center_at,omitempty"`
}

type CosmeticPaintDropShadow struct {
	OffsetX float64     `json:"x_offset" bson:"x_offset"`
	OffsetY float64     `json:"y_offset" bson:"y_offset"`
	Radius  float64     `json:"radius" bson:"radius"`
	Color   utils.Color `json:"color" bson:"color"`
}

type CosmeticPaintText struct {
	// Weight multiplier for the text. Default is 9
	Weight uint8 `json:"weight" bson:"weight"`
	// Shadows applied to the text
	Shadow []CosmeticPaintDropShadow `json:"shadow,omitempty" bson:"shadow,omitempty"`
	// Text tranformation
	Transform CosmeticPaintTextTransform `json:"transform,omitempty" bson:"transform,omitempty"`
	// Text stroke
	Stroke *CosmeticPaintStroke `json:"stroke,omitempty" bson:"stroke,omitempty"`
	// (css) font variant property. non-standard
	Variant string `json:"variant" bson:"variant"`
}

type CosmeticPaintStroke struct {
	// Stroke color
	Color utils.Color `json:"color" bson:"color"`
	// Stroke width
	Width float64 `json:"width" bson:"width"`
}

type CosmeticPaintTextTransform string

const (
	CosmeticPaintTextTransformUppercase CosmeticPaintTextTransform = "uppercase"
	CosmeticPaintTextTransformLowercase CosmeticPaintTextTransform = "lowercase"
)

type CosmeticPaintFlair struct {
	// The kind of sprite
	Kind CosmeticPaintFlairKind `json:"kind" bson:"kind"`
	// The width of the sprite
	Width float64 `json:"width" bson:"width"`
	// The height of the sprite
	Height float64 `json:"height" bson:"height"`
	// Base64-encoded image or vector data
	Data string `json:"data" bson:"data"`
}

type CosmeticPaintFlairKind string

const (
	CosmeticPaintSpriteKindImage  CosmeticPaintFlairKind = "IMAGE"
	CosmeticPaintSpriteKindVector CosmeticPaintFlairKind = "VECTOR"
	CosmeticPaintSpriteKindText   CosmeticPaintFlairKind = "TEXT"
)
