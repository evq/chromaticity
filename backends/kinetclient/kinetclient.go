package kinetclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib" 
	"github.com/evq/go-kinet"
	"github.com/lucasb-eyer/go-colorful"
	"image/color"
	"strconv"
  "reflect"
)

type Backend struct {
	PowerSupplies []kinet.PowerSupply `json:"powerSupplies"`
  NextColor map[string][]color.Color `json:"-"`
  CurrentColor map[string][]color.Color `json:"-"`
}

type KinetLight struct {
	chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Fixture    *kinet.Fixture      `json:"-"`
  Backend `json:"-"`
  NextColor *color.Color `json:"-"`
}

func (k KinetLight) SetColor(c colorful.Color) {
	// Gamma correction
	r, g, b := c.Clamped().LinearRgb()
  *k.NextColor = color.RGBA{
    uint8(r * 255),
    uint8(g * 255),
    uint8(b * 255),
    255,
  }
}

func (b Backend) Sync() {
	pses := b.PowerSupplies
	for i := range pses {
    ps := pses[i]
    if !reflect.DeepEqual(b.NextColor[ps.Mac], b.CurrentColor[ps.Mac]) {
      ps.SendColors(b.NextColor[ps.Mac])
      currentColors := b.CurrentColor[ps.Mac]
      for j := range currentColors {
        currentColors[j] = b.NextColor[ps.Mac][j]
      }
    }
  }
}

func (k KinetLight) GetState() *chromaticity.State {
	return k.LightState
}

func (b Backend) GetType() string {
	return "kinet"
}

func (b Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, &b)
	pses := b.PowerSupplies

	for i := range pses {
		ps := pses[i]
    b.NextColor[ps.Mac] = make([]color.Color, len(ps.Fixtures))
    b.CurrentColor[ps.Mac] = make([]color.Color, len(ps.Fixtures))
    AddPSLights(l, &ps, b.NextColor[ps.Mac])
	}
}

func (b Backend) DiscoverLights(l *chromaticity.LightResource) {
	powerSupplies := kinet.Discover()

	for i := range powerSupplies {
		ps := powerSupplies[i]
		exists := false
		for j := range l.Lights {
			light := interface{}(*l.Lights[j])
			switch v := (light).(type) {
			case *KinetLight:
				syncedps := v.Fixture.PS
				if ps.IP == syncedps.IP {
          // Scanning resets currently acive color
					//chromaticity.SendState(l.Lights[j])
					exists = true
					break
				}
			default:
				continue
			}
		}

		if !exists {
      b.NextColor[ps.Mac] = make([]color.Color, len(ps.Fixtures))
      b.CurrentColor[ps.Mac] = make([]color.Color, len(ps.Fixtures))
			AddPSLights(l, ps, b.NextColor[ps.Mac])
		}
	}
}

func AddPSLights(l *chromaticity.LightResource, ps *kinet.PowerSupply, nextColor []color.Color) {
	for j := range ps.Fixtures {
		ps.Fixtures[j].PS = ps
	}

	kinetLights := LightsFrom(ps, nextColor)

	ids := []string{}
	for k := range kinetLights {
		id := strconv.Itoa(len(l.Lights) + 1)
		light := &kinetLights[k]
		l.Lights[id] = light
		ids = append(ids, id)
	}

	l.Groups[strconv.Itoa(len(l.Groups)+1)] = chromaticity.Group{
		chromaticity.GroupInfo{ids, ps.Name},
		&chromaticity.NewState().ColorState,
		true,
	}
}

func LightsFrom(ps *kinet.PowerSupply, nextColor []color.Color) []chromaticity.Light {
	lights := []chromaticity.Light{}
	for i := range ps.Fixtures {
		k := KinetLight{}
		k.Fixture = ps.Fixtures[i]
    k.NextColor = &nextColor[k.Fixture.Channel]

		if k.Fixture.Color == nil {
			k.Fixture.Color = colorful.LinearRgb(0.0, 0.0, 0.0)
		}
		r, g, b, _ := k.Fixture.Color.RGBA()
		c := colorful.LinearRgb(float64(r)/65535.0, float64(g)/65535.0, float64(b)/65535.0)

		k.LightState = chromaticity.NewState()
		k.LightState.SetColor(c)

		k.Type = "Philips ColorKinetics"
		k.Name = ps.Name + " Serial:" + k.Fixture.Serial // default, make this changable
		k.ModelId = ps.Type
		k.SwVersion = ps.FWVersion

		//lights[strconv.Itoa(len(lights) + 1)] = &k
		lights = append(lights, &k)
	}
	return lights
}
