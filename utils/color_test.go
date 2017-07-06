package utils_test

import (
	"fmt"
	"github.com/evq/chromaticity/utils"
	"github.com/lucasb-eyer/go-colorful"
)

func ExampleRgbToRgbx_amber() {
	rgb, x := utils.RgbToRgbx(colorful.Color{1.0, 1.0, 0.0}, colorful.Color{1.0, 1.0, 0})
	fmt.Println(rgb.R, rgb.G, rgb.B, x)
	// Output: foo
}

func ExampleRgbToRgbx_red() {
	rgb, x := utils.RgbToRgbx(colorful.Color{1.0, 0.0, 0.0}, colorful.Color{1.0, 1.0, 0})
	fmt.Println(rgb.R, rgb.G, rgb.B, x)
	// Output: foo
}
func ExampleRgbToRgbw_red() {
	rgb, x := utils.RgbToRgbw(colorful.Color{1.0, 0.0, 0.0}, 1000000/5000)
	fmt.Println(rgb.R, rgb.G, rgb.B, x)
	// Output: foo
}

func ExampleRgbToRgbw_white() {
	rgb, x := utils.RgbToRgbw(colorful.Color{1.0, 1.0, 0.9}, 1000000/5000)
	fmt.Println(rgb.R, rgb.G, rgb.B, x)
	// Output: foo
}
