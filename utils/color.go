package utils

import (
	"fmt"
	"strconv"
)

// Color is a utility to efficiently format colors into a 32-bit integer with alpha channel support
type Color int32

// ParseHex parses a hex string into a Color
func (c Color) ParseHex(v string, alpha uint8) (Color, error) {
	val, err := strconv.ParseInt(v, 16, 32)
	if err != nil {
		return 0, err
	}

	t := int32(val)
	c = c.SetRed(uint8((t << 16) & 255)).
		SetGreen(uint8((t << 8) & 255)).
		SetBlue(uint8(t & 255)).
		SetAlpha(alpha)

	return c, nil
}

// ParseHex parses red, green, blue into a Color
func (c Color) ParseRGB(r, g, b, a uint8) Color {
	return c.SetRed(r).SetGreen(g).SetBlue(b).SetAlpha(a)
}

func (c Color) Sum() int32 {
	return int32(c)
}

// SetRed sets the red channel of the color
func (c Color) SetRed(v uint8) Color {
	return c | (Color(v) << 24)
}

// SetGreen sets the green channel of the color
func (c Color) SetGreen(v uint8) Color {
	return c | (Color(v) << 16)
}

// SetBlue sets the blue channel of the color
func (c Color) SetBlue(v uint8) Color {
	return c | (Color(v) << 8)
}

// SetAlpha sets the alpha channel of the color
func (c Color) SetAlpha(v uint8) Color {
	return c | Color(v)
}

// SetRGB sets the red, green, and blue channels of the color
func (c Color) SetRGB(r, g, b uint8) Color {
	return c.SetRed(r).SetGreen(g).SetBlue(b)
}

// SetRGBA sets the red, green, blue, and alpha channels of the color
func (c Color) SetRGBA(r, g, b, a uint8) Color {
	return c.SetRed(r).SetGreen(g).SetBlue(b).SetAlpha(a)
}

// GetRed returns the red channel of the color
func (c Color) GetRed() uint8 {
	return uint8((c >> 24) & 0xFF)
}

// SetGreen sets the green channel of the color
func (c Color) GetGreen() uint8 {
	return uint8((c >> 16) & 0xFF)
}

// GetBlue returns the blue channel of the color
func (c Color) GetBlue() uint8 {
	return uint8((c >> 8) & 0xFF)
}

// GetAlpha returns the alpha channel of the color
func (c Color) GetAlpha() uint8 {
	return uint8(c & 0xFF)
}

// ToRGB returns the red, green, and blue channels of the color
func (c Color) ToRGB() [3]uint8 {
	return [3]uint8{
		c.GetRed(),
		c.GetGreen(),
		c.GetBlue(),
	}
}

// ToRGBA returns the red, green, blue, and alpha channels of the color
func (c Color) ToRGBA() [4]uint8 {
	return [4]uint8{
		c.GetRed(),
		c.GetGreen(),
		c.GetBlue(),
		c.GetAlpha(),
	}
}

// ToHex returns the hex representation of the color
func (c Color) ToHex(alpha bool) string {
	if alpha {
		return fmt.Sprintf("#%02x%02x%02x%02x", c.GetRed(), c.GetGreen(), c.GetBlue(), c.GetAlpha())
	} else {
		return fmt.Sprintf("#%02x%02x%02x", c.GetRed(), c.GetGreen(), c.GetBlue())
	}
}
