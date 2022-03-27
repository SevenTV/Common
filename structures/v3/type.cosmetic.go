package structures

import (
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Cosmetic struct {
	ID       primitive.ObjectID   `json:"id" bson:"_id"`
	Kind     CosmeticKind         `json:"kind" bson:"kind"`
	Priority int                  `json:"priority" bson:"priority"`
	Name     string               `json:"name" bson:"name"`
	UserIDs  []primitive.ObjectID `json:"users" bson:"user_ids"`
	Users    []*User              `json:"user_objects" bson:"user_objects,skip,omitempty"`
	Data     bson.Raw             `json:"data" bson:"data"`

	// User Relationals
	Selected bool `json:"selected" bson:"selected,skip,omitempty"`
}

func (c *Cosmetic) decode(i interface{}) interface{} {
	if err := bson.Unmarshal(c.Data, i); err != nil {
		logrus.WithError(err).Error("cosmetic, failed to decode data")
	}
	return i
}

func (c *Cosmetic) ReadBadge() *CosmeticDataBadge {
	b := c.decode(&CosmeticDataBadge{}).(*CosmeticDataBadge)
	b.ID = c.ID
	return b
}
func (c *Cosmetic) ReadPaint() *CosmeticDataPaint {
	b := c.decode(&CosmeticDataPaint{}).(*CosmeticDataPaint)
	b.ID = c.ID
	return b
}

type CosmeticKind string

const (
	CosmeticKindBadge        CosmeticKind = "BADGE"
	CosmeticKindNametagPaint CosmeticKind = "PAINT"
)

type CosmeticDataBadge struct {
	ID      primitive.ObjectID `json:"id" bson:"-"`
	Tooltip string             `json:"tooltip" bson:"tooltip"`
	Misc    bool               `json:"misc,omitempty" bson:"misc"`
}

type CosmeticDataPaint struct {
	ID primitive.ObjectID `json:"id" bson:"-"`
	// The function used to generate the paint (i.e gradients or an image)
	Function CosmeticPaintFunction `json:"function" bson:"function"`
	// The default color of the paint
	Color *int32 `json:"color" bson:"color"`
	// Gradient stops, a list of positions and colors
	Stops []CosmeticPaintGradientStop `json:"stops" bson:"stops"`
	// Whether or not the gradient repeats outside its original area
	Repeat bool `json:"repeat" bson:"repeat"`
	// Gradient angle in degrees
	Angle int32 `json:"angle" bson:"angle"`
	// Shape of a radial gradient, when the paint is of RADIAL_GRADIENT type
	Shape string `json:"shape,omitempty" bson:"shape,omitempty"`
	// URL of an image, when the paint is of BACKGROUND_IMAGE type
	ImageURL string `json:"image_url,omitempty" bson:"image_url,omitempty"`
	// A list of drop shadows. There may be any amount, which can be stacked onto each other
	DropShadows []CosmeticPaintDropShadow `json:"drop_shadows,omitempty" bson:"drop_shadows,omitempty"`
}

type CosmeticPaintFunction string

const (
	CosmeticPaintFunctionLinearGradient CosmeticPaintFunction = "linear-gradient"
	CosmeticPaintFunctionRadialGradient CosmeticPaintFunction = "radial-gradient"
	CosmeticPaintFunctionImageURL       CosmeticPaintFunction = "url"
)

type CosmeticPaintGradientStop struct {
	At    float64 `json:"at" bson:"at"`
	Color int32   `json:"color" bson:"color"`
}

type CosmeticPaintDropShadow struct {
	OffsetX float64 `json:"x_offset" bson:"x_offset"`
	OffsetY float64 `json:"y_offset" bson:"y_offset"`
	Radius  float64 `json:"radius" bson:"radius"`
	Color   int32   `json:"color" bson:"color"`
}

type CosmeticPaintAnimation struct {
	Speed     int32                            `json:"speed" bson:"speed"`
	Keyframes []CosmeticPaintAnimationKeyframe `json:"keyframes" bson:"keyframes"`
}

type CosmeticPaintAnimationKeyframe struct {
	At float64 `json:"at" bson:"at"`
	X  float64 `json:"x" bson:"x"`
	Y  float64 `json:"y" bson:"y"`
}