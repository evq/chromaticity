package chromaticity

import (
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
  "github.com/lucasb-eyer/go-colorful"
	"net/http"
  "fmt"
)

type Group struct {
	GroupInfo
	State    *ColorState `json:"action"`
	ReadOnly bool        `json:"-"`
  Lights   []*Light    `json:"-"`
	//Scenes
}

type GroupInfo struct {
	LightIDs []string `json:"lights"`
	Name     string   `json:"name"`
}

func NewGroup(l *LightResource, i GroupInfo, s *ColorState, ro bool) (*Group) {
  g := Group{}
  g.GroupInfo = i
  g.State = s
  g.ReadOnly = ro
  g.Lights = make([]*Light, len(g.LightIDs))
	for i := range g.LightIDs {
		id := g.LightIDs[i]
    g.Lights[i] = l.Lights[id]
  }
  return &g
}

func (g Group) SetColor(c colorful.Color) {
  for l := range g.Lights {
    (*g.Lights[l]).SetColor(c)
  }
}

func (g Group) SetColors(c []colorful.Color) {
  for l := range g.Lights {
    (*g.Lights[l]).SetColor(c[l])
  }
}

func (g Group) GetState() (*State) {
  state := State{}
  state.ColorState = g.State
  return &state
}

func (g Group) GetNumPixels() uint16 {
  return uint16(len(g.LightIDs))
}

func (l LightResource) listGroups(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Groups)
}

func (l LightResource) updateGroupState(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	// Special case, hidden group of all lights
  // FIXME doesn't work for effects
	var group Group
	if id == "0" {
		ids := []string{}
		for k, _ := range l.Lights {
			ids = append(ids, k)
		}
		group = *NewGroup(
      &l,
			GroupInfo{ids, "All Lights"},
			NewState().ColorState,
			true,
		)
	} else {
		group = l.Groups[id]
	}
	if len(group.Name) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Group could not be found.")
		return
	}

	UpdateColorState(group.State, request)

	for i := range group.LightIDs {
		id := group.LightIDs[i]
    state := (*l.Lights[id]).GetState()
		UpdateColorState(state.ColorState, request)
    if state.EffectRoutine != nil {
      fmt.Println("Canceling")
      if !state.EffectRoutine.Done {
        state.EffectRoutine.Done = true
        state.EffectRoutine.Signal <- true
        close(state.EffectRoutine.Signal)
      }
      state.EffectRoutine = nil
      fmt.Println("done canceling")
    }
	}

  grouplight := Light(group)
  SendState(&grouplight)

  if group.State.EffectRoutine != nil {
    for i := range group.LightIDs {
      id := group.LightIDs[i]
      state := (*l.Lights[id]).GetState()
      state.EffectRoutine = group.State.EffectRoutine
    }
  }

	response.WriteEntity(group)
}

func (l LightResource) RegisterGroupsApi(container *restful.Container) {
	utils.RegisterApis(
		container,
		"/groups",
		"Manage Groups",
		l._RegisterGroupsApi,
	)
}

func (l LightResource) _RegisterGroupsApi(ws *restful.WebService) {
	ws.Route(ws.GET("/").To(l.listGroups).
		Doc("list all groups").
		Operation("listGroups"))

	ws.Route(ws.PUT("/{group-id}/action").To(l.updateGroupState).
		Doc("modify a groups's state").
		Operation("updateGroupState").
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Reads(ColorState{}))
}
