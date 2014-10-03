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
