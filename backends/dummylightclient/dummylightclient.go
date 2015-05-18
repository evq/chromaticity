package dummylightclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
)

type Backend struct {
	Devices []DummyLightbulb `json:"lights"`
}

type DummyLightbulb struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"-"`
	Type       string              `json:"type"`
}

func (k DummyLightbulb) SetColor(c colorful.Color) {}

func (k DummyLightbulb) SetColors(c []colorful.Color) {}

func (b *Backend) Sync() {}

func (z DummyLightbulb) GetNumPixels() uint16 {
	return 1
}

func (z DummyLightbulb) GetInfo() *chromaticity.LightInfo {
	return z.LightInfo
}

func (z DummyLightbulb) GetState() *chromaticity.State {
	return z.LightState
}

func (z DummyLightbulb) GetType() string {
	return "dummylight"
}

func (b Backend) GetType() string {
	return "dummylight"
}

func (b Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, &b)

	// The bottom code allows for adding of bulbs during runtime
	for x := range b.Devices {
		light := DummyLightbulb{}
		light.LightInfo = &chromaticity.LightInfo{}
		light.LightState = chromaticity.NewState()
		light.Type = b.Devices[x].Type
		var castlight chromaticity.Light = &light
		l.Lights[strconv.Itoa(len(l.Lights)+1)] = &castlight
	}
}

func (b Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
