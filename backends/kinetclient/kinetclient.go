package kinetclient

import (
	"encoding/json"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/go-kinet"
	"github.com/lucasb-eyer/go-colorful"
	"image/color"
	"strconv"
	"time"
)

type Backend struct {
	PowerSupplies []kinet.PowerSupply          `json:"powerSupplies"`
	NextColor     map[string][]*colorful.Color `json:"-"`
	CurrentColor  map[string][]*colorful.Color `json:"-"`
}

type KinetLight struct {
	*chromaticity.LightInfo
	LightState *chromaticity.State `json:"state"`
	Fixture    *kinet.Fixture      `json:"-"`
	Backend    `json:"-"`
	NextColor  *colorful.Color `json:"-"`
}

func (k KinetLight) SetColor(c colorful.Color) {
	*k.NextColor = c
}

func (k KinetLight) SetColors(c []colorful.Color) {
	k.SetColor(c[0])
}

func (b Backend) Sync() {
	pses := b.PowerSupplies
	for i := range pses {
		ps := pses[i]
		go b.PSSync(&ps)
	}
}

func (b Backend) PSSync(ps *kinet.PowerSupply) {
	for {
		eq := true

		arr := make([]color.Color, len(b.NextColor[ps.Mac]))
		for e := range arr {
			// FIXME, switch to utils implementation of Clamp
			r, g, bb := (*b.NextColor[ps.Mac][e]).Clamped().RGB255()
			arr[e] = color.RGBA{r, g, bb, 0xFF}
			if *b.NextColor[ps.Mac][e] != *b.CurrentColor[ps.Mac][e] {
				eq = false
			}
		}

		if !eq {
			ps.SendColors(arr)
			currentColors := b.CurrentColor[ps.Mac]
			for j := range currentColors {
				*currentColors[j] = *b.NextColor[ps.Mac][j]
			}
		}
		time.Sleep(15 * time.Millisecond)
	}
}

func (k KinetLight) GetNumPixels() uint16 {
	return 1
}

func (k KinetLight) GetInfo() *chromaticity.LightInfo {
	return k.LightInfo
}

func (k KinetLight) GetState() *chromaticity.State {
	return k.LightState
}

func (k KinetLight) GetType() string {
	return "kinet"
}

func (b Backend) GetType() string {
	return "kinet"
}

func (b Backend) ImportLights(l *chromaticity.LightResource, from []byte) {
	json.Unmarshal(from, &b)
	pses := b.PowerSupplies

	for i := range pses {
		ps := pses[i]
		b.NextColor[ps.Mac] = make([]*colorful.Color, len(ps.Fixtures))
		b.CurrentColor[ps.Mac] = make([]*colorful.Color, len(ps.Fixtures))
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
			if b.NextColor == nil {
				b.NextColor = make(map[string][]*colorful.Color)
			}
			if b.CurrentColor == nil {
				b.CurrentColor = make(map[string][]*colorful.Color)
			}
			b.NextColor[ps.Mac] = make([]*colorful.Color, len(ps.Fixtures))
			for i := range b.NextColor[ps.Mac] {
				b.NextColor[ps.Mac][i] = &colorful.Color{}
			}
			b.CurrentColor[ps.Mac] = make([]*colorful.Color, len(ps.Fixtures))
			for i := range b.CurrentColor[ps.Mac] {
				b.CurrentColor[ps.Mac][i] = &colorful.Color{}
			}
			AddPSLights(l, ps, b.NextColor[ps.Mac])
			go b.PSSync(ps)
		}
	}
}

func AddPSLights(l *chromaticity.LightResource, ps *kinet.PowerSupply, nextColor []*colorful.Color) {
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

	l.Groups[strconv.Itoa(len(l.Groups)+1)] = *chromaticity.NewGroup(
		l,
		chromaticity.GroupInfo{ids, ps.Name, chromaticity.Luminaire},
		chromaticity.NewState().ColorState,
		true,
	)
}

func LightsFrom(ps *kinet.PowerSupply, nextColor []*colorful.Color) []chromaticity.Light {
	lights := []chromaticity.Light{}
	for i := range ps.Fixtures {
		k := KinetLight{}
		k.Fixture = ps.Fixtures[i]
		k.NextColor = nextColor[i]

		if k.Fixture.Color == nil {
			k.Fixture.Color = colorful.LinearRgb(0.0, 0.0, 0.0)
		}
		r, g, b, _ := k.Fixture.Color.RGBA()
		c := colorful.LinearRgb(float64(r)/65535.0, float64(g)/65535.0, float64(b)/65535.0)

		k.LightState = chromaticity.NewState()
		k.LightState.SetColor(c)

		k.LightInfo = &chromaticity.LightInfo{}
		k.Type = chromaticity.Ex_Color_Light
		k.Name = ps.Name + " Serial:" + k.Fixture.Serial // default, make this changable
		k.ModelId = "ColorKinetics " + ps.Type
		k.SwVersion = ps.FWVersion

		k.PointSymbol = make(map[string]string, 8)
		for l := 1; l < 9; l++ {
			k.PointSymbol[strconv.Itoa(l)] = "none"
		}

		lights = append(lights, &k)
	}
	return lights
}
