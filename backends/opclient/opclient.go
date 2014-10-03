package opclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/kellydunn/go-opc"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
)

type Backend struct {
	Servers []OPCServer `json:"servers"`
}

type OPCLight struct {
	chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Chan       Channel             `json:"-"`
}

type OPCServer struct {
	Name     string      `json:"name"`
	Host     string      `json:"host"`
	Port     string      `json:"port"`
	Channels []Channel   `json:"channels"`
	Client   *opc.Client `json:"-"`
}

type Channel struct {
	ID        uint8      `json:"id"`
	NumPixels uint16     `json:"numPixels"`
	Server    *OPCServer `json:"-"`
}

func (k OPCLight) SendColor(c colorful.Color) {
	if k.Chan.Server.Client == nil {
		k.Chan.Server.Client = opc.NewClient()
		hostPort := k.Chan.Server.Host + ":" + k.Chan.Server.Port
		k.Chan.Server.Client.Connect("tcp", hostPort)
	}
	r, g, b := c.Clamped().LinearRgb()

	msg := opc.NewMessage(k.Chan.ID)
	msg.SetLength(k.Chan.NumPixels * 3)
	for i := 0; i < int(k.Chan.NumPixels); i++ {
		msg.SetPixelColor(
			i,
			uint8(r*255),
			uint8(g*255),
			uint8(b*255),
		)
	}
	k.Chan.Server.Client.Send(msg)
}

func (k OPCLight) GetState() *chromaticity.State {
	return k.LightState
}

func (b Backend) GetType() string {
	return "opc"
}

func (b Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, &b)

	for i := range b.Servers {
		server := b.Servers[i]
		for j := range server.Channels {
			light := OPCLight{}
			light.Chan = server.Channels[j]
			light.Chan.Server = &server
			light.Type = "OPC Light"
			light.Name = server.Name + " Chan:" + strconv.Itoa(int(light.Chan.ID))
			light.ModelId = ":D"
			light.SwVersion = "0"

			light.LightState = chromaticity.NewState()

			id := strconv.Itoa(len(l.Lights) + 1)

			var castlight chromaticity.Light = &light
			l.Lights[id] = &castlight
		}
	}
}

func (b Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
