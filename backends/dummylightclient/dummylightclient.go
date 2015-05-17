package dummylightclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
)

type Backend struct {
	Devices [1]FakeLightbulb `json:"-"`
}

type FakeLightbulb struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
}

type FakeLightbulbDev struct {
}

func (k FakeLightbulb) SetColor(c colorful.Color) {}

func (k FakeLightbulb) SetColors(c []colorful.Color) {}

func (b *Backend) Sync() {}

func (z FakeLightbulb) GetNumPixels() uint16 {
	return 1
}

func (z FakeLightbulb) GetInfo() *chromaticity.LightInfo {
	return z.LightInfo
}

func (z FakeLightbulb) GetState() *chromaticity.State {
	return z.LightState
}

func (z FakeLightbulb) GetType() string {
	return "FakeLight"
}

func (b Backend) GetType() string {
	return "FakeLight"
}

func (b Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, &b)
	// b.Devices[0] = FakeLightbulb{}
	light := FakeLightbulb{}
	light.LightInfo = &chromaticity.LightInfo{}
	light.LightState = chromaticity.NewState()
	var castlight chromaticity.Light = &light
	l.Lights[strconv.Itoa(len(l.Lights)+1)] = &castlight
}

func (b Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
