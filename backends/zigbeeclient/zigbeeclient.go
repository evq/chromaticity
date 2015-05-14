package zigbeeclient

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/go-apron"
	"github.com/evq/go-zigbee"
	"github.com/evq/go-zigbee/gateways/embercli"
	"github.com/lucasb-eyer/go-colorful"
	"reflect"
	"strconv"
	"time"
)

type Backend struct {
	Gateway *embercli.EmberCliGateway `json:"gateway"`
	Aprondb string                    `json:"aprondb"`
	Devices []ZigbeeLightDev          `json:"-"`
}

type ZigbeeLight struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Device     *ZigbeeLightDev     `json:"-"`
}

type ZigbeeLightDev struct {
	zigbee.ZigbeeDevice
	CurrentColor   colorful.Color      `json:"-"`
	NextColor      colorful.Color      `json:"-"`
	LightState     *chromaticity.State `json:"state"`
	Effect         string              `json:"-"`
	TransitionTime uint16              `json:"-"`
	EndpointId     uint8               `json:"-"`
	Type           string              `json:"-"`
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
	var err error
	if b.Gateway != nil {
		err = b.Gateway.Reconnect()
	} else {
		b.Gateway = &embercli.EmberCliGateway{}
		err = b.Gateway.Connect("127.0.0.1:4901")
	}
	if err != nil {
		// FIXME add auto reconnect
		log.Error("[chromaticity/backends/zigbee] Unable to connect to zigbee gateway: " + b.Gateway.Address)
	} else {
		for {
			for i := range b.Devices {
				dev := b.Devices[i]
				state := dev.LightState
				if !reflect.DeepEqual(dev.NextColor, dev.CurrentColor) {
					if !state.On {
						b.Gateway.MoveToLightLevelWOnOff(dev.ZigbeeDevice, dev.EndpointId, 0x00, state.TransitionTime)
					} else {
						b.Gateway.MoveToLightLevelWOnOff(dev.ZigbeeDevice, dev.EndpointId, state.Bri, state.TransitionTime)
						b.Gateway.Send()

						if dev.Type == chromaticity.Ex_Color_Light {
							switch state.Colormode {
							case "xy":
								b.Gateway.MoveToXY(dev.ZigbeeDevice, dev.EndpointId, uint16(state.Xy[0]*65535), uint16(state.Xy[1]*65535), state.TransitionTime)
							case "hs":
								b.Gateway.MoveToHueSat(dev.ZigbeeDevice, dev.EndpointId, uint8(state.Hue/257), dev.LightState.Sat, dev.LightState.TransitionTime)
							case "ct":
								b.Gateway.MoveToColorTemp(dev.ZigbeeDevice, dev.EndpointId, dev.LightState.Ct, dev.LightState.TransitionTime)
							}
						}
					}
					b.Gateway.Send()
					b.Devices[i].CurrentColor = dev.NextColor

					if state.Effect != "none" && dev.Type == chromaticity.Ex_Color_Light {
						b.Gateway.Loop(dev.ZigbeeDevice, dev.EndpointId, state.Hue, state.TransitionTime)
						b.Gateway.Send()
					}
				}
				if dev.Type == chromaticity.Ex_Color_Light {
					if dev.Effect != state.Effect || dev.TransitionTime != state.TransitionTime {
						if state.Effect != "none" {
							b.Gateway.Loop(dev.ZigbeeDevice, dev.EndpointId, state.Hue, state.TransitionTime)
						} else {
							b.Gateway.StopLoop(dev.ZigbeeDevice, dev.EndpointId, state.Hue, state.TransitionTime)
						}
						b.Gateway.Send()
						b.Devices[i].Effect = dev.LightState.Effect
						b.Devices[i].TransitionTime = dev.LightState.TransitionTime
					}
				}
			}
			time.Sleep(time.Duration(10) * time.Millisecond)
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

	if b.Aprondb == "" {
		b.Aprondb = "/database/apron.db"
	}

	db, err := apron.Open(b.Aprondb)
	if err != nil {
		log.Error("[chromaticity/backends/zigbee] Unable to open apron database: " + b.Aprondb)
		return
	}

	devices := db.GetZigbeeDevices()
	b.Devices = make([]ZigbeeLightDev, len(devices), len(devices))

	for i := range devices {
		b.Devices[i].ZigbeeDevice = devices[i]

		var endpoint *zigbee.Endpoint
		// Assume single endpoint per device
		for b.Devices[i].EndpointId, endpoint = range devices[i].Endpoints {
			break
		}

		light := ZigbeeLight{}
		light.Device = &b.Devices[i]

		light.LightInfo = &chromaticity.LightInfo{}

		if endpoint.DeviceType == zigbee.ZLLExtendedColorLight {
			light.Type = chromaticity.Ex_Color_Light
			light.Device.Type = chromaticity.Ex_Color_Light
		} else {
			light.Type = chromaticity.Dimmable_Light
			light.Device.Type = chromaticity.Dimmable_Light
		}

		light.PointSymbol = make(map[string]string, 8)
		for k := 1; k < 9; k++ {
			light.PointSymbol[strconv.Itoa(k)] = "none"
		}
		basicinfo := (*endpoint).InClusters[zigbee.BasicCluster].Attributes
		light.Name = devices[i].Name
		light.ModelId = basicinfo[zigbee.ModelId]
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
