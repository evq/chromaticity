package chromaticity

import (
  "time"
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
	"github.com/lucasb-eyer/go-colorful"
	"net/http"
  //"fmt"
)

type State struct {
	Reachable bool   `json:"reachable"`
	Colormode string `json:"colormode"`
	ColorState
  EffectRoutine chan bool `json:"-"`
}

type ColorState struct {
	Alert          string    `json:"alert"`
	Bri            uint8     `json:"bri"`
	Ct             uint16    `json:"ct"`
	Effect         string    `json:"effect"`
  EffectSpread   float64   `json:"effectspread"`
	Hue            uint16    `json:"hue"`
	On             bool      `json:"on"`
	TransitionTime uint16    `json:"transitiontime"`
	Sat            uint8     `json:"sat"`
	Xy             []float64 `json:"xy"`
}

type LightResource struct {
  Lights map[string]*Light `json:"lights"`
  Groups map[string]Group `json:"groups"`
  *ConfigInfo `json:"config"`
}

type Light interface {
	SetColor(c colorful.Color)
	SetColors(c []colorful.Color)
	GetState() (s *State)
  GetNumPixels() (p uint16)
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
	//cs := (*light).GetState().ColorState
	//cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
	//request.ReadEntity(&cs)

  //fmt.Println(cs)

	//UpdateColorState(light, cs)
	UpdateColorState((*light).GetState(), request)
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

func UpdateColorState(dest *State, req *restful.Request) {
  cs := (*dest).ColorState
	cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
  cs.Effect = "none"
	req.ReadEntity(&cs)

	mode := _UpdateColorState(&(*dest).ColorState, cs)
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

	if s.On != state.On {
		state.On = s.On
	}
	// alert

  state.Effect = s.Effect

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

func SendState(l *Light) {
  s := (*l).GetState()
  if s.EffectRoutine != nil {
    s.EffectRoutine <- true
    s.EffectRoutine = nil
  }

  switch s.Effect {
    case "colorloop":
      s.EffectRoutine = make(chan bool)
      go HsvLoop(s.EffectRoutine, l)
    case "hsvloop":
      s.EffectRoutine = make(chan bool)
      go HsvLoop(s.EffectRoutine, l)
    case "hclloop":
      s.EffectRoutine = make(chan bool)
      go HclLoop(s.EffectRoutine, l)
    default:
      (*l).SetColor(s.GetColor())
  }
}

// A hsv based rainbow fade
func HsvLoop(done chan bool, light *Light) {
  s := (*light).GetState()
  h, c, l := s.GetColor().Hsv()
  for {
    select {
      case <- done:
          return
      default:
        h = h + 1.00
        if h >= 360.0 {
          h = 0.0
        }
        j := h
        colors := make([]colorful.Color, (*light).GetNumPixels())
        for i := range colors {
          colors[i] = colorful.Hsv(j,c,l)
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
      case <- done:
          return
      default:
        h = h + 1.00
        if h >= 360.0 {
          h = 0.0
        }
        j := h
        colors := make([]colorful.Color, (*light).GetNumPixels())
        for i := range colors {
          colors[i] = colorful.Hcl(j,c,l)
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
