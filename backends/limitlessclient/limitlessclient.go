package limitlessclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/go-limitless"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
)

type Backend struct {
	Controllers []limitless.LimitlessController `json:"controllers"`
}

type LimitlessLight struct {
	chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Group       limitless.LimitlessGroup   `json:"-"`
}

func (k LimitlessLight) SendColor(c colorful.Color) {
  k.Group.SendColor(c.Clamped())
}

func (k LimitlessLight) GetState() *chromaticity.State {
	return k.LightState
}

func (b Backend) GetType() string {
	return "limitless"
}

func (b Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, &b)

	for i := range b.Controllers {
		controller := b.Controllers[i]
    ids := []string{}
		for j := range controller.Groups {
			light := LimitlessLight{}
			light.Group = controller.Groups[j]
			light.Group.Controller = &controller
			light.Type = "LimitlessLED Light"
			light.Name = controller.Name + ": " + light.Group.Name
			light.ModelId = light.Group.Type
			light.SwVersion = "v3.0"

			light.LightState = chromaticity.NewState()

			id := strconv.Itoa(len(l.Lights) + 1)
      ids = append(ids, id)

			var castlight chromaticity.Light = &light
			l.Lights[id] = &castlight
		}

    l.Groups[strconv.Itoa(len(l.Groups)+1)] = chromaticity.Group{
      chromaticity.GroupInfo{ids, controller.Name},
      &chromaticity.NewState().ColorState,
      true,
    }
	}
}

func (b Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
