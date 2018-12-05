package tplinkclient

import (
	"encoding/json"
	"strconv"

	"github.com/appnaconda/tplink"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/lucasb-eyer/go-colorful"
)

type Backend struct {
	Clients []TPLinkClient `json:"clients"`
}

type TPLinkLight struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Client     *TPLinkClient       `json:"-"`
}

type TPLinkClient struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (b *Backend) GetType() string {
	return "tplink"
}

func (b *Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, b)
	for i := range b.Clients {
		light := TPLinkLight{}
		light.LightInfo = &chromaticity.LightInfo{}
		light.Type = chromaticity.On_Off_Plug_Light
		light.PointSymbol = make(map[string]string, 8)
		for k := 1; k < 9; k++ {
			light.PointSymbol[strconv.Itoa(k)] = "none"
		}
		light.Name = b.Clients[i].Name
		light.ModelId = "TPLink HS100"
		light.SwVersion = "0"
		light.LightState = chromaticity.NewState()
		light.Client = &b.Clients[i]

		id := strconv.Itoa(len(l.Lights) + 1)

		var castlight chromaticity.Light = &light
		l.Lights[id] = &castlight
	}
}

func (b *Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}

func (b *Backend) Sync() {
	return
}

func (t *TPLinkLight) SetColor(c colorful.Color) {
	on := true
	if c.R == 0 && c.G == 0 && c.B == 0 {
		on = false
	}

	plug := tplink.NewHS100(t.Client.Address)
	if on {
		plug.TurnOn()
	} else {
		plug.TurnOff()
	}
}

func (t *TPLinkLight) SetColors(c []colorful.Color) {
	return
}

func (t *TPLinkLight) GetInfo() (i *chromaticity.LightInfo) {
	return t.LightInfo
}

func (t *TPLinkLight) GetState() (s *chromaticity.State) {
	return t.LightState
}

func (t *TPLinkLight) GetNumPixels() (p uint16) {
	return 1
}

func (t *TPLinkLight) GetType() string {
	return "tplink"
}
