package opclient

import (
	"encoding/json"
	"fmt"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/chromaticity/utils"
	"github.com/kellydunn/go-opc"
	"github.com/lucasb-eyer/go-colorful"
	"reflect"
	"strconv"
	"time"
)

type Backend struct {
	Servers []OPCServer `json:"servers"`
}

type OPCLight struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Chan       *Channel            `json:"-"`
}

type OPCServer struct {
	Name        string      `json:"name"`
	Address     string      `json:"address"`
	RefreshRate uint8       `json:"refreshrate"`
	Type        string      `json:"type"`
	White       uint16      `json:"white"`
	Gamma       float64     `json:"gamma"`
	Channels    []Channel   `json:"channels"`
	Client      *opc.Client `json:"-"`
}

type Channel struct {
	ID            uint8            `json:"id"`
	NumPixels     uint16           `json:"numPixels"`
	Server        *OPCServer       `json:"-"`
	CurrentColors []colorful.Color `json:"-"`
	NextColors    []colorful.Color `json:"-"`
}

func (o OPCLight) SetColor(c colorful.Color) {
	for i := range o.Chan.NextColors {
		o.Chan.NextColors[i] = c
	}
}

func (o OPCLight) SetColors(c []colorful.Color) {
	for i := range c {
		o.Chan.NextColors[i] = c[i]
	}
}

func (b *Backend) Sync() {
	for s := range b.Servers {
		server := &b.Servers[s]
		go server.Sync()
	}
}

func (server *OPCServer) Sync() {
	for {
		for c := range server.Channels {
			channel := server.Channels[c]
			if !reflect.DeepEqual(channel.NextColors, channel.CurrentColors) {
				msg := opc.NewMessage(channel.ID)
				if server.Type == "RGBW" || server.Type == "RGBA" {
					msg.SetLength(2 * channel.NumPixels * 3)
				} else {
					msg.SetLength(channel.NumPixels * 3)
				}

				for i := 0; i < int(channel.NumPixels); i++ {
					p := channel.NextColors[i]
					utils.Clamp(&p)
					if server.Type == "RGBW" || server.Type == "RGBA" {
						var rgb colorful.Color
						var x float64
						if server.Type == "RGBW" {
							rgb, x = utils.RgbToRgbw(p, server.White)
						} else if server.Type == "RGBA" {
							rgb, x = utils.RgbToRgbx(p, colorful.Color{1.0, 0.8, 0})
						}
						msg.SetPixelColor(
							2*i,
							uint8(utils.Linearize(rgb.R, server.Gamma)*255),
							uint8(utils.Linearize(rgb.G, server.Gamma)*255),
							uint8(utils.Linearize(rgb.B, server.Gamma)*255),
						)
						msg.SetPixelColor(
							2*i+1,
							uint8(utils.Linearize(x, server.Gamma)*255),
							0,
							0,
						)
					} else {
						msg.SetPixelColor(
							i,
							uint8(utils.Linearize(p.R, server.Gamma)*255),
							uint8(utils.Linearize(p.G, server.Gamma)*255),
							uint8(utils.Linearize(p.B, server.Gamma)*255),
						)
					}
					channel.CurrentColors[i] = channel.NextColors[i]
				}

				server.Connect()

				for {
					if server.Client != nil {
						err := server.Client.Send(msg)
						if err != nil {
							fmt.Print("ERROR!:")
							fmt.Print(err.Error())
							server.Client = nil
							server.Connect()
							continue
						}
					}
					break
				}
			}
		}
		time.Sleep(time.Duration(1000.0/float64(server.RefreshRate)) * time.Millisecond)
	}
}

func (server *OPCServer) Connect() {
	if server.Client == nil {
		server.Client = opc.NewClient()
		err := server.Client.Connect("tcp", server.Address)
		if err != nil {
			fmt.Print("ERROR!:")
			fmt.Print(err.Error())
			server.Client = nil
		}
	}
}

func (o OPCLight) GetNumPixels() uint16 {
	return o.Chan.NumPixels
}

func (o OPCLight) GetInfo() *chromaticity.LightInfo {
	return o.LightInfo
}

func (o OPCLight) GetState() *chromaticity.State {
	return o.LightState
}

func (o OPCLight) GetType() string {
	return "opc"
}

func (b *Backend) GetType() string {
	return "opc"
}

func (b *Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, b)

	for i := range b.Servers {
		if b.Servers[i].Type == "" {
			b.Servers[i].Type = "RGB"
		}
		if b.Servers[i].Gamma == 0 {
			b.Servers[i].Gamma = 2.2
		}
		server := b.Servers[i]
		for j := range server.Channels {
			light := OPCLight{}
			light.LightInfo = &chromaticity.LightInfo{}
			light.Chan = &server.Channels[j]
			light.Chan.Server = &server
			light.Chan.CurrentColors = make([]colorful.Color, light.Chan.NumPixels)
			light.Chan.NextColors = make([]colorful.Color, light.Chan.NumPixels)
			if server.Type == "W" {
				light.Type = chromaticity.Dimmable_Light
			} else {
				light.Type = chromaticity.Ex_Color_Light
			}
			light.PointSymbol = make(map[string]string, 8)
			for k := 1; k < 9; k++ {
				light.PointSymbol[strconv.Itoa(k)] = "none"
			}
			if len(server.Channels) == 1 {
				light.Name = server.Name
			} else {
				light.Name = server.Name + " C" + strconv.Itoa(int(light.Chan.ID))
			}
			light.ModelId = "OPC-" + server.Type
			light.SwVersion = "0"

			light.LightState = chromaticity.NewState()

			id := strconv.Itoa(len(l.Lights) + 1)

			var castlight chromaticity.Light = &light
			l.Lights[id] = &castlight
		}
	}
}

func (b *Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
