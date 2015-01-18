package opclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/kellydunn/go-opc"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
  "reflect"
  "fmt"
  "time"
)

type Backend struct {
	Servers []OPCServer `json:"servers"`
}

type OPCLight struct {
	chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Chan       *Channel             `json:"-"`
}

type OPCServer struct {
	Name     string      `json:"name"`
	Host     string      `json:"host"`
	Port     string      `json:"port"`
	RefreshRate     uint8      `json:"refreshrate"`
	Channels []Channel   `json:"channels"`
	Client   *opc.Client `json:"-"`
}

type Channel struct {
	ID        uint8      `json:"id"`
	NumPixels uint16     `json:"numPixels"`
	Server    *OPCServer `json:"-"`
  CurrentColors []colorful.Color `json:"-"`
  NextColors []colorful.Color    `json:"-"`
}

func (k OPCLight) SetColor(c colorful.Color) {
  for i := range k.Chan.NextColors {
    k.Chan.NextColors[i] = c
  }
}

func (k OPCLight) SetColors(c []colorful.Color) {
  for i := range c {
    k.Chan.NextColors[i] = c[i]
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
        if server.Client == nil {
          server.Client = opc.NewClient()
          hostPort := server.Host + ":" + server.Port
          err := server.Client.Connect("tcp", hostPort)
          if err != nil {
            fmt.Print("ERROR!:")
            fmt.Print(err.Error())
          }
        }

        msg := opc.NewMessage(channel.ID)
        msg.SetLength(channel.NumPixels * 3)

        for i := 0; i < int(channel.NumPixels); i++ {
          r, g, b := channel.NextColors[i].Clamped().LinearRgb()
          msg.SetPixelColor(
            i,
            uint8(r*255),
            uint8(g*255),
            uint8(b*255),
          )
          channel.CurrentColors[i] = channel.NextColors[i]
        }
        server.Client.Send(msg)
      }
    }
    time.Sleep(time.Duration(1000.0 / float64(server.RefreshRate)) * time.Millisecond)
  }
}

func (k OPCLight) GetNumPixels() uint16 {
  return k.Chan.NumPixels
}

func (k OPCLight) GetState() *chromaticity.State {
	return k.LightState
}

func (b *Backend) GetType() string {
	return "opc"
}

func (b *Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, b)

	for i := range b.Servers {
		server := b.Servers[i]
		for j := range server.Channels {
			light := OPCLight{}
			light.Chan = &server.Channels[j]
			light.Chan.Server = &server
      light.Chan.CurrentColors = make([]colorful.Color, light.Chan.NumPixels)
      light.Chan.NextColors = make([]colorful.Color, light.Chan.NumPixels)
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

func (b *Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
