package utils

import (
	"github.com/lucasb-eyer/go-colorful"
	"math"
)

// McCamy cubic approximation
func ToMirads(c colorful.Color) uint16 {
	x, y, _ := c.Xyy()
	n := (x - 0.3320) / (y - 0.1858)
	return uint16(1000000 / (499.0*math.Pow(n, 3) + 3525.0*math.Pow(n, 2) - 6823.3*n + 5520.33))
}

func cubic(a float64, b float64, c float64, d float64, x float64) float64 {
	return (a*math.Pow(x, 3) + b*math.Pow(x, 2) + c*x + d)
}

func FromMirads(temp uint16, bri uint8) colorful.Color {
	b := float64(bri) / 255.0
	k := 1000000.0 / float64(temp)

	if k < 1667.0 {
		k = 1667.0
	} else if k > 25000 {
		k = 25000.0
	}

	var x float64
	var y float64

	if k < 4000.0 {
		x = cubic(-0.2661239, -0.2343580, 0.8776956, 0.179910, 1000/k)
	} else {
		x = cubic(-3.0258469, 2.1070379, 0.2226347, 0.240390, 1000/k)
	}

	if k < 2222.0 {
		y = cubic(-1.1063814, -1.34811020, 2.18555832, -0.20219683, x)
	} else if k < 4000.0 {
		y = cubic(-0.9549476, -1.37418593, 2.09137015, -0.16748867, x)
	} else {
		y = cubic(3.0817580, -5.87338670, 3.75112997, -0.37001483, x)
	}
	c := colorful.Xyy(x, y, 1.0)
	c.R *= b
	c.G *= b
	c.B *= b
	return c
}

func WhiteCmpt(c colorful.Color, w colorful.Color) float64 {
	_, sat, _ := c.Hsv()
	diff := w.DistanceCIE94(c)
	if diff > 1.0 {
		diff = 1.0
	}

	return (1.0 - diff) * (1.0 - math.Pow(sat, 5))
}

func RgbToRgbw(c colorful.Color, mir uint16) (rgb colorful.Color, w float64) {
	white := FromMirads(mir, 255)
	Clamp(&white)

	_, _, v := c.Hsv()
	Maximize(&c)

	w = WhiteCmpt(c, white)

	white.R = w * white.R
	white.G = w * white.G
	white.B = w * white.B

	rgb.R = c.R - white.R
	rgb.G = c.G - white.G
	rgb.B = c.B - white.B

	max := math.Max(rgb.R, math.Max(rgb.G, rgb.B))

	rgb.R += (1.0-max)*c.R
	rgb.G += (1.0-max)*c.G
	rgb.B += (1.0-max)*c.B

	rgb.R *= v
	rgb.G *= v
	rgb.B *= v
	w *= v

	return
}

/// Linear ///
//////////////
// https://github.com/lucasb-eyer/go-colorful/blob/master/colors.go
// http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/
// http://www.brucelindbloom.com/Eqn_RGB_to_XYZ.html

//func Linearize(v float64) float64 {
//if v <= 0.04045 {
//return v / 12.92
//}
//return 1.055 * math.Pow(v, 1.0/2.4) - 0.055
//}

func Linearize(v float64, gamma float64) float64 {
	return math.Pow(v, gamma)
}

func Clamp(c *colorful.Color) {
	max := math.Max(c.R, math.Max(c.G, c.B))
	if max > 1.0 {
		c.R /= max
		c.G /= max
		c.B /= max
	}

	if c.R < 0.0 {
		c.R = 0
	}
	if c.G < 0.0 {
		c.G = 0
	}
	if c.B < 0.0 {
		c.B = 0
	}
	return
}

func Maximize(c * colorful.Color) {
	max := math.Max(c.R, math.Max(c.G, c.B))
	if max < 1.0 {
		c.R /= max
		c.G /= max
		c.B /= max
	}
	return
}

// hsv2rgb_rainbow converted to operate on floats 0..1
// from https://github.com/FastLED/FastLED/blob/master/hsv2rgb.cpp
func Hsv2Rainbow(h float64, s float64, v float64) colorful.Color {
	h = h / 360.0

	// offset = h & 0x1F
	// offset8 = offset * 8
	offset := 8.0 * math.Mod(h, 0.125)
	third := offset / 3.0
	twothird := offset * 2.0 / 3.0


	r := 0.0
	g := 0.0
	b := 0.0
	if h < 0.5 { // ! (h & 0x80)
		if h < 0.25 { // ! (h & 0x40)
			if h < 0.125 { // ! (h & 0x20)
				r = 1.0 - third
				g = third
				b = 0
			} else {
				// Y1
				r = 0.666 // 171
				g = 0.333 + third //85
				b = 0
				// Y2
				//r = 0.666 + third
				//g = 0.333 + twothird
				//b = 0
			}
		} else {
			if h < 0.375 { // ! (h & 0x20)
				// Y1
				r = 0.666 - twothird
				g = 0.666 + third
				b = 0
				// Y2
				//r = 1.0 - offset
				//g = 1.0
				//b = 0.0
			} else {
				r = 0
				g = 1.0 - third
				b = third
			}
		}
	} else {
		if h < 0.75 { // ! (h & 0x40)
			if h < 0.625 { // ! (h & 20)
				r = 0
				g = 0.666 - twothird
				b = 0.333 + twothird
			} else {
				r = third
				g = 0
				b = 1.0 - third
			}
		} else {
			if h < 0.875 { // ! (h & 20)
				r = 0.333 + third
				g = 0
				b = 0.666 - third
			} else {
				r = 0.666 + third
				g = 0
				b = 0.333 - third
			}
		}
	}
	g *= 0.5

	r *= v
	g *= v
	b *= v

	return colorful.Color{r,g,b}
}
