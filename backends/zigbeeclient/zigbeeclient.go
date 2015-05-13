package zigbeeclient

import (
	"encoding/json"
	//"fmt"
	chromaticity "github.com/evq/chromaticity/lib"
	//"github.com/evq/chromaticity/utils"
	"github.com/evq/go-zigbee"
	"github.com/evq/go-zigbee/gateways/embercli"
	"github.com/lucasb-eyer/go-colorful"
	"reflect"
	"strconv"
	"time"
)

const ENDPOINT = 0x0b

type Backend struct {
	Gateway *embercli.EmberCliGateway `json:"gateway"`
	Devices []ZigbeeLightDev `json:"devices"`
}

type ZigbeeLight struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Device *ZigbeeLightDev `json:"-"`
}

type ZigbeeLightDev struct {
	zigbee.ZigbeeDevice
	CurrentColor colorful.Color `json:"-"`
	NextColor    colorful.Color `json:"-"`
	LightState *chromaticity.State `json:"state"`
	Effect string `json:"-"`
	TransitionTime uint16 `json:"-"`
}

func (k ZigbeeLight) SetColor(c colorful.Color) {
		k.Device.NextColor = c
}

func (k ZigbeeLight) SetColors(c []colorful.Color) {
}

func (b *Backend) Sync() {
	go b._Sync()
}

func (b *Backend) _Sync() {
	if b.Gateway != nil {
		b.Gateway.Reconnect()
		for {
			for i := range b.Devices {
				dev := b.Devices[i]
				state := dev.LightState
				if !reflect.DeepEqual(dev.NextColor, dev.CurrentColor) {
					if !state.On {
						b.Gateway.MoveToLightLevelWOnOff(dev.ZigbeeDevice, ENDPOINT, 0x00, state.TransitionTime)
					} else {
						b.Gateway.MoveToLightLevelWOnOff(dev.ZigbeeDevice, ENDPOINT, state.Bri, state.TransitionTime)
						b.Gateway.Send()
						switch state.Colormode {
							case "xy":
								b.Gateway.MoveToXY(dev.ZigbeeDevice, ENDPOINT, uint16(state.Xy[0] * 65535), uint16(state.Xy[1] * 65535), state.TransitionTime)
							case "hs":
								b.Gateway.MoveToHueSat(dev.ZigbeeDevice, ENDPOINT, uint8(state.Hue/257), dev.LightState.Sat, dev.LightState.TransitionTime)
							case "ct":
								b.Gateway.MoveToColorTemp(dev.ZigbeeDevice, ENDPOINT, dev.LightState.Ct, dev.LightState.TransitionTime)
						}
					}
					b.Gateway.Send()
					b.Devices[i].CurrentColor = dev.NextColor

					if state.Effect != "none" {
							b.Gateway.Loop(dev.ZigbeeDevice, ENDPOINT, state.Hue, state.TransitionTime)
					}
					b.Gateway.Send()
				}
				if dev.Effect != state.Effect || dev.TransitionTime != state.TransitionTime {
					if state.Effect != "none" {
							b.Gateway.Loop(dev.ZigbeeDevice, ENDPOINT, state.Hue, state.TransitionTime)
					} else {
							b.Gateway.StopLoop(dev.ZigbeeDevice, ENDPOINT, state.Hue, state.TransitionTime)
					}
					b.Gateway.Send()
					b.Devices[i].Effect = dev.LightState.Effect
					b.Devices[i].TransitionTime = dev.LightState.TransitionTime
				}
			}
			//time.Sleep(time.Duration(1000.0/float64(server.RefreshRate)) * time.Millisecond)
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
	}
}

func (z ZigbeeLight) GetNumPixels() uint16 {
	return 1
}

func (z ZigbeeLight) GetInfo() *chromaticity.LightInfo {
	return z.LightInfo
}

func (z ZigbeeLight) GetState() *chromaticity.State {
	return z.LightState
}

func (z ZigbeeLight) GetType() string {
	return "zigbee"
}

func (b *Backend) GetType() string {
	return "zigbee"
}

func (b *Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, b)

	for i := range b.Devices {
		light := ZigbeeLight{}

		light.Device = &b.Devices[i]

		light.LightInfo = &chromaticity.LightInfo{}
		light.Type = "Extended color light"
		light.PointSymbol = make(map[string]string, 8)
		for k:= 1; k < 9; k++ {
			light.PointSymbol[strconv.Itoa(k)] = "none"
		}
		light.Name = b.Devices[i].Name
		light.ModelId = "FIXME"
		light.SwVersion = "0"
		light.LightState = chromaticity.NewState()
		b.Devices[i].LightState = light.LightState

		id := strconv.Itoa(len(l.Lights) + 1)

		var castlight chromaticity.Light = &light
		l.Lights[id] = &castlight
	}
}

func (b *Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
