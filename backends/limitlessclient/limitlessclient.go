package limitlessclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/go-limitless"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
  "time"
  "fmt"
)

type Backend struct {
	Controllers []Controller `json:"controllers"`
}

type Controller struct {
  limitless.LimitlessController
  CurrentColors map[int]*colorful.Color `json:"-"`
  NextColors map[int]*colorful.Color `json:"-"`
}

type LimitlessLight struct {
	chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Group       *limitless.LimitlessGroup   `json:"-"`
  NextColor *colorful.Color `json:"-"`
}

func (k LimitlessLight) SetColor(c colorful.Color) {
  (*k.NextColor) = c
}

func (k LimitlessLight) SetColors(c []colorful.Color) {
  k.SetColor(c[0])
}

func (b *Backend) Sync() {
  for c:= range b.Controllers {
    controller := &b.Controllers[c]
    go controller.Sync()
  }
}

func (c *Controller) Sync() {
  for {
    for g := range c.Groups {
      group := c.Groups[g]
      if *c.NextColors[group.Id] != *c.CurrentColors[group.Id] {
        err := group.SendColor(*c.NextColors[group.Id])
        if err != nil {
          fmt.Println(err.Error())
        }
        *c.CurrentColors[group.Id] = *c.NextColors[group.Id]
      } else {
        time.Sleep(10 * time.Millisecond)
      }
    }
  }
}

func (k LimitlessLight) GetNumPixels() uint16 {
  return 1
}

func (k LimitlessLight) GetState() *chromaticity.State {
	return k.LightState
}

func (b *Backend) GetType() string {
	return "limitless"
}

func (b *Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, b)

	for i := range b.Controllers {
		controller := &b.Controllers[i]
    ids := []string{}
    controller.NextColors = make(map[int]*colorful.Color)
    controller.CurrentColors = make(map[int]*colorful.Color)
		for j := range controller.Groups {
      controller.NextColors[controller.Groups[j].Id] = &colorful.Color{}
      controller.CurrentColors[controller.Groups[j].Id] = &colorful.Color{}
			light := LimitlessLight{}
      light.NextColor = controller.NextColors[controller.Groups[j].Id]
			light.Group = &controller.Groups[j]
			light.Group.Controller = &controller.LimitlessController
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

    l.Groups[strconv.Itoa(len(l.Groups)+1)] = *chromaticity.NewGroup(
      l,
      chromaticity.GroupInfo{ids, controller.Name},
      chromaticity.NewState().ColorState,
      true,
    )
	}
}

func (b *Backend) DiscoverLights(l *chromaticity.LightResource) {
	return
}
