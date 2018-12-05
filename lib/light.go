package chromaticity

import (
	"net/http"
	"time"

	"github.com/evq/chromaticity/utils"
	"github.com/evq/go-restful"
	"github.com/lucasb-eyer/go-colorful"
)

type State struct {
	Reachable bool `json:"reachable"`
	*ColorState
}

type ColorState struct { // Corresponds to "action" in Hue API
	Alert          string    `json:"alert"`
	Bri            uint8     `json:"bri"`
	BriInc         int16     `json:"bri_inc,omitempty"`
	Ct             uint16    `json:"ct"`
	Effect         string    `json:"effect"`
	EffectSpread   float64   `json:"effectspread,omitempty"`
	Hue            uint16    `json:"hue"`
	On             bool      `json:"on"`
	TransitionTime uint16    `json:"transitiontime,omitempty"`
	Sat            uint8     `json:"sat"`
	Xy             []float64 `json:"xy"`
	Colormode      string    `json:"colormode"`
	*EffectRoutine
}

type EffectRoutine struct {
	Signal chan bool `json:"-"`
	Done   bool      `json:"-"`
}

type LightResource struct {
	Lights      map[string]*Light    `json:"lights"`
	Groups      map[string]*Group    `json:"groups"`
	Schedules   map[string]*Schedule `json:"schedules"`
	*ConfigInfo `json:"config"`
}

type Light interface {
	SetColor(c colorful.Color)
	SetColors(c []colorful.Color)
	GetInfo() (i *LightInfo)
	GetState() (s *State)
	GetNumPixels() (p uint16)
	GetType() string
}

// Light Types
const (
	Ex_Color_Light    = "Extended color light"
	Dimmable_Light    = "Dimmable light"
	On_Off_Plug_Light = "On/Off plug-in unit"
)

type LightInfo struct {
	// Api def limits these strings to certain length, FIXME if doesn't work
	Type      string `json:"type,omitempty"`
	Name      string `json:"name,omitempty"`
	ModelId   string `json:"modelid,omitempty"`
	SwVersion string `json:"swversion,omitempty"`
	// Api def claims unimplemented atm, FIXME if doesn't work
	PointSymbol map[string]string `json:"pointsymbol"`
}

type LuxLight struct {
	State *LuxState `json:"state"`
	*LightInfo
}

// Omit color fields for white only lights
type LuxState struct {
	*State
	Ct        bool `json:"ct,omitempty"`
	Hue       bool `json:"hue,omitempty"`
	Sat       bool `json:"sat,omitempty"`
	Xy        bool `json:"xy,omitempty"`
	Colormode bool `json:"colormode,omitempty"`
}

func NewState() *State {
	s := State{}
	s.ColorState = &ColorState{}
	s.Alert = "none"
	s.Alert = "none"
	s.Effect = "none"
	s.On = true
	s.Reachable = true
	s.Colormode = "xy"
	s.Xy = []float64{0.0, 0.0}
	return &s
}

func (l LightResource) findLight(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("light-id")
	light := l.Lights[id]
	if light == nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Light could not be found.")
		return
	}
	response.WriteEntity(WrapLight(light))
}

func (l LightResource) listLights(request *restful.Request, response *restful.Response) {
	response.WriteEntity(WrapLights(l.Lights))
}

func (l LightResource) updateLightState(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("light-id")
	light := l.Lights[id]
	if light == nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Light could not be found.")
		return
	}
	cs := (*light).GetState().ColorState
	last_color := cs.GetColor()
	UpdateColorState(cs, request)

	SendState(light, last_color)

	response.WriteEntity(WrapLight(light))
}

func (l LightResource) RegisterLightsApi(container *restful.Container) {
	utils.RegisterApis(
		container,
		"/api/{api_key}/lights",
		"Manage Lights",
		l._RegisterLightsApi,
	)
}

func (l LightResource) _RegisterLightsApi(ws *restful.WebService) {
	ws.Route(ws.GET("/").To(l.listLights).
		Doc("list all lights").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("listLights"))

	ws.Route(ws.GET("/{light-id}").To(l.findLight).
		Doc("get a light").
		Operation("findLight").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Param(ws.PathParameter("light-id", "identifier of the light").DataType("int")))

	ws.Route(ws.PUT("/{light-id}/state").To(l.updateLightState).
		Doc("modify a light's state").
		Operation("updateLightState").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Param(ws.PathParameter("light-id", "identifier of the light").DataType("int")).
		Reads(ColorState{}))
}

func WrapLights(l map[string]*Light) map[string]interface{} {
	lights := make(map[string]interface{})
	for k, v := range l {
		lights[k] = WrapLight(v)
	}
	return lights
}

func WrapLight(l *Light) interface{} {
	info := (*l).GetInfo()
	if (*info).Type == "Dimmable light" {
		luxs := LuxState{}
		luxs.State = (*l).GetState()
		lux := LuxLight{}
		lux.State = &luxs
		lux.LightInfo = info
		return lux
	} else {
		return l
	}
}

func (state *State) SetColor(c colorful.Color) {
	x, y, bri := c.Xyy()
	state.Bri = uint8(bri * 255.0)
	switch state.Colormode {
	case "xy":
		state.Xy = []float64{x, y}
	case "hs":
		h, s, _ := c.Hsv()
		state.Hue = uint16(h / 360.0 * 65535.0)
		state.Sat = uint8(s * 255.0)
	case "ct":
		state.Ct = utils.ToMirads(c)
	}
}

func (state *ColorState) GetColor() (c colorful.Color) {
	if !state.On {
		return colorful.Xyz(0.0, 0.0, 0.0)
	}
	switch state.Colormode {
	case "xy":
		return colorful.Xyy(state.Xy[0], state.Xy[1], float64(state.Bri)/255.0)
	case "ct":
		return utils.FromMirads(state.Ct, state.Bri)
	case "hs":
		return colorful.Hsv(
			float64(state.Hue)*360.0/65535.0,
			float64(state.Sat)/255.0,
			float64(state.Bri)/255.0,
		)
	}
	return
}

func UpdateColorState(dest *ColorState, req *restful.Request) {
	cs := (*dest)
	cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
	req.ReadEntity(&cs)

	mode := _UpdateColorState(dest, cs)
	if len(mode) != 0 {
		dest.Colormode = mode
	} else {
		cs = ColorState{}
		cs.Xy = []float64{0.0, 0.0}
		cs.Effect = "none"
		src := cs
		req.ReadEntity(&src)
		mode = _UpdateColorState(&cs, src)
		if len(mode) != 0 {
			dest.Colormode = mode
		}
	}
}

// This function is designed to take an updated copy of the colorstate
// from the light as second argument to allow partial PUT / PATCH.
//
// An alternative to using pointers for fields ala
// https://github.com/google/go-github/issues/19 which somehow seems
// like no fun at all :)
func _UpdateColorState(state *ColorState, s ColorState) string {
	mode := ""

	if len(s.Xy) == 2 && len(state.Xy) == 2 {
		if s.Xy[0] != state.Xy[0] || s.Xy[1] != state.Xy[1] {
			mode = "xy"
			state.Xy[0] = s.Xy[0]
			state.Xy[1] = s.Xy[1]
		}
	}

	if s.Ct != state.Ct {
		if len(mode) == 0 {
			mode = "ct"
		}
		state.Ct = s.Ct
	}

	if s.Hue != state.Hue {
		if len(mode) == 0 {
			mode = "hs"
		}
		state.Hue = s.Hue
	}

	if s.Sat != state.Sat {
		if len(mode) == 0 {
			mode = "hs"
		}
		state.Sat = s.Sat
	}

	if s.Bri != state.Bri {
		state.Bri = s.Bri
	}

	if s.BriInc != state.BriInc {
		if s.BriInc > -254 && s.BriInc < 254 {
			state.Bri = uint8(int16(state.Bri) + s.BriInc)
		}
	}

	if s.On != state.On {
		state.On = s.On
	}
	// alert

	if s.Effect != state.Effect {
		state.Effect = s.Effect
	}

	if s.TransitionTime != state.TransitionTime {
		state.TransitionTime = s.TransitionTime
	}

	if s.EffectSpread != state.EffectSpread {
		if s.EffectSpread >= 0.0 {
			state.EffectSpread = s.EffectSpread
		}
	}

	return mode
}

func SendState(l *Light, last_color colorful.Color) {
	s := (*l).GetState()
	if s.EffectRoutine != nil {
		if !s.EffectRoutine.Done {
			s.EffectRoutine.Done = true
			s.EffectRoutine.Signal <- true
			close(s.EffectRoutine.Signal)
		}
		s.EffectRoutine = nil
	}

	if (*l).GetType() != "zigbee" {
		switch s.Effect {
		case "colorloop":
			s.EffectRoutine = &EffectRoutine{make(chan bool), false}
			go HsvLoop(s.EffectRoutine.Signal, l)
		case "hsvloop":
			s.EffectRoutine = &EffectRoutine{make(chan bool), false}
			go HsvLoop(s.EffectRoutine.Signal, l)
		case "hclloop":
			s.EffectRoutine = &EffectRoutine{make(chan bool), false}
			go HclLoop(s.EffectRoutine.Signal, l)
		case "rainbow":
			s.EffectRoutine = &EffectRoutine{make(chan bool), false}
			go RainbowLoop(s.EffectRoutine.Signal, l)
		default:
			s.EffectRoutine = &EffectRoutine{make(chan bool), false}
			go BlendColor(s.EffectRoutine, l, last_color)
		}
	} else {
		(*l).SetColor(s.GetColor())
	}
}

func BlendColor(e *EffectRoutine, light *Light, last_color colorful.Color) {
	s := (*light).GetState()
	next_color := s.GetColor()
	utils.Clamp(&next_color)
	for i := 0; i < 100*int(s.TransitionTime); i++ {
		select {
		case <-e.Signal:
			break
		default:
			color := last_color.BlendRgb(next_color, float64(i)/(100.0*float64(s.TransitionTime)))
			//fmt.Println(color)
			(*light).SetColor(color)
			time.Sleep(time.Millisecond)
		}
	}
	(*light).SetColor(next_color)
	e.Done = true
	return
}

// A hsv based rainbow fade
func RainbowLoop(done chan bool, light *Light) {
	s := (*light).GetState()
	h, c, l := s.GetColor().Hsv()
	for {
		select {
		case <-done:
			return
		default:
			h = h + 1.00
			if h >= 360.0 {
				h = 0.0
			}
			j := h
			colors := make([]colorful.Color, (*light).GetNumPixels())
			for i := range colors {
				colors[i] = utils.Hsv2Rainbow(j, c, l)
				j = j + ((*light).GetState().EffectSpread * 360.0 / float64((*light).GetNumPixels()))
				if j >= 360.0 {
					j = j - 360.0
				}
			}
			(*light).SetColors(colors)
			time.Sleep((10 + time.Duration(s.TransitionTime)) * time.Millisecond)
		}
	}
}

// A hsv based rainbow fade
func HsvLoop(done chan bool, light *Light) {
	s := (*light).GetState()
	h, c, l := s.GetColor().Hsv()
	for {
		select {
		case <-done:
			return
		default:
			h = h + 1.00
			if h >= 360.0 {
				h = 0.0
			}
			j := h
			colors := make([]colorful.Color, (*light).GetNumPixels())
			for i := range colors {
				colors[i] = colorful.Hsv(j, c, l)
				j = j + ((*light).GetState().EffectSpread * 360.0 / float64((*light).GetNumPixels()))
				if j >= 360.0 {
					j = j - 360.0
				}
			}
			(*light).SetColors(colors)
			time.Sleep((10 + time.Duration(s.TransitionTime)) * time.Millisecond)
		}
	}
}

// A hcl based rainbow fade
func HclLoop(done chan bool, light *Light) {
	s := (*light).GetState()
	h, c, l := s.GetColor().Hcl()
	for {
		select {
		case <-done:
			return
		default:
			h = h + 1.00
			if h >= 360.0 {
				h = 0.0
			}
			j := h
			colors := make([]colorful.Color, (*light).GetNumPixels())
			for i := range colors {
				colors[i] = colorful.Hcl(j, c, l)
				j = j + ((*light).GetState().EffectSpread * 360.0 / float64((*light).GetNumPixels()))
				if j >= 360.0 {
					j = j - 360.0
				}
			}
			(*light).SetColors(colors)
			time.Sleep((10 + time.Duration(s.TransitionTime)) * time.Millisecond)
		}
	}
}
