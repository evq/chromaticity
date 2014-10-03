package chromaticity

import (
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
	"github.com/lucasb-eyer/go-colorful"
	"net/http"
)

type State struct {
	Reachable bool   `json:"reachable"`
	Colormode string `json:"colormode"`
	ColorState
}

type ColorState struct {
	Alert          string    `json:"alert"`
	Bri            uint8     `json:"bri"`
	Ct             uint16    `json:"ct"`
	Effect         string    `json:"effect"`
	Hue            uint16    `json:"hue"`
	On             bool      `json:"on"`
	TransitionTime uint16    `json:"transitiontime"`
	Sat            uint8     `json:"sat"`
	Xy             []float64 `json:"xy"`
}

type LightResource struct {
	Lights map[string]*Light
	Groups map[string]Group
}

type Light interface {
	SendColor(c colorful.Color)
	GetState() (s *State)
}

type LightInfo struct {
	// Api def limits these strings to certain length, FIXME if doesn't work
	Type      string `json:"type"`
	Name      string `json:"name"`
	ModelId   string `json:"modelid"`
	SwVersion string `json:"swversion"`
	// Api def claims unimplemented atm, FIXME if doesn't work
	PointSymbol map[string]string `json:"-"`
}

func NewState() *State {
	s := State{}
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
	response.WriteEntity(light)
}

func (l LightResource) listLights(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Lights)
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
	cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
	request.ReadEntity(&cs)
	UpdateColorState(light, cs)
	SendState(light)
	response.WriteEntity(light)
}

func (l LightResource) RegisterLightsApi(container *restful.Container) {
	utils.RegisterApis(
		container,
		"/lights",
		"Manage Lights",
		l._RegisterLightsApi,
	)
}

func (l LightResource) _RegisterLightsApi(ws *restful.WebService) {
	ws.Route(ws.GET("/").To(l.listLights).
		Doc("list all lights").
		Operation("listLights"))

	ws.Route(ws.GET("/{light-id}").To(l.findLight).
		Doc("get a light").
		Operation("findLight").
		Param(ws.PathParameter("light-id", "identifier of the light").DataType("int")))

	ws.Route(ws.PUT("/{light-id}/state").To(l.updateLightState).
		Doc("modify a light's state").
		Operation("updateLightState").
		Param(ws.PathParameter("light-id", "identifier of the light").DataType("int")).
		Reads(ColorState{}))
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

func (state *State) GetColor() (c colorful.Color) {
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

func UpdateColorState(l *Light, s ColorState) {
	state := (*l).GetState()
	mode := _UpdateColorState(&state.ColorState, s)
	if len(mode) != 0 {
		state.Colormode = mode
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

	if s.On != state.On {
		state.On = s.On
	}
	// alert
	// effect
	// transitiontime
	return mode
}

func SendState(l *Light) {
	(*l).SendColor((*l).GetState().GetColor())
}
