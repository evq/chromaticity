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
		y = cubic(-1.1063814, -1.34811020, 2.18555832, 0.20219683, x)
	} else if k < 4000.0 {
		y = cubic(-0.9549476, -1.37418593, 2.09137015, -0.16748867, x)
	} else {
		y = cubic(3.0817580, -5.87338670, 3.75112997, -0.37001483, x)
	}
	return colorful.Xyy(x, y, b)
}

func WhiteCmpt(c colorful.Color, w colorful.Color) float64 {
	_,sat,_ := c.Hsv()
	diff := w.DistanceCIE94(c)
	if diff > 1.0 {
		diff = 1.0
	}

	mag := c.R + c.G + c.B / 3.0
	if mag > 1.0 {
		mag = 1.0
	}

	if mag <= 0.0 {
		return 0.0
	} else {
		mag = 1.0 / mag
	}

	return (1.0 - diff) * (1.0 - math.Pow(sat,3)) * mag
}

func RgbToRgbw(c colorful.Color, mir uint16) (rgb colorful.Color, w float64) {
	white := FromMirads(mir, 255)
	w = WhiteCmpt(c, white)
	white.R = w * white.R
	white.G = w * white.G
	white.B = w * white.B

	rgb.R = c.R - white.R
	rgb.G = c.G - white.G
	rgb.B = c.B - white.B

	rgb.R = rgb.R + (1.0 - math.Max(rgb.R, math.Max(rgb.G, rgb.B))) * c.R
	rgb.G = rgb.G + (1.0 - math.Max(rgb.R, math.Max(rgb.G, rgb.B))) * c.G
	rgb.B = rgb.B + (1.0 - math.Max(rgb.R, math.Max(rgb.G, rgb.B))) * c.B

	return
}

/// Linear ///
//////////////
// https://github.com/lucasb-eyer/go-colorful/blob/master/colors.go
// http://www.sjbrown.co.uk/2004/05/14/gamma-correct-rendering/
// http://www.brucelindbloom.com/Eqn_RGB_to_XYZ.html

func Linearize(v float64) float64 {
	if v <= 0.04045 {
			return v / 12.92
	}
	return 1.055 * math.Pow(v, 1.0/2.4) - 0.055
}
