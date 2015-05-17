package chromaticity

import (
	"encoding/json"
	"fmt"
	"github.com/evq/chromaticity/utils"
	"github.com/evq/go-restful"
	"github.com/lucasb-eyer/go-colorful"
	"io"
	"log"
	"net/http"
	"strings"
)

type Group struct {
	GroupInfo
	State    *ColorState `json:"action"`
	ReadOnly bool        `json:"-"`
	Lights   []*Light    `json:"-"`
	//Scenes
}

type GroupInfo struct {
	LightIDs  []string `json:"lights"`
	Name      string   `json:"name"`
	GroupType string   `json:"type"`
}

func NewGroup(l *LightResource, i GroupInfo, s *ColorState, ro bool) *Group {
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

func (g Group) GetInfo() *LightInfo {
	info := LightInfo{}
	return &info
}

func (g Group) GetState() *State {
	state := State{}
	state.ColorState = g.State
	return &state
}

func (g Group) GetNumPixels() uint16 {
	return uint16(len(g.LightIDs))
}

func (g Group) GetType() string {
	return "group"
}

func (l LightResource) listGroups(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Groups)
}

func (l LightResource) createGroup(request *restful.Request, response *restful.Response) {
	var lights []string
	dec := json.NewDecoder(strings.NewReader(request.PathParameter("lights")))

	type Message struct {
		Light string
	}
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		lights = append(lights, m.Light)
	}
	thisLen := fmt.Sprintf("%d", len(l.Groups))
	thisState := &((*l.Lights[lights[0]]).GetState().ColorState)
	ginfo := GroupInfo{lights,
		request.PathParameter("name"),
		request.PathParameter("type")}
	l.Groups[thisLen] = Group{
		ginfo,
		thisState, // Copies the state from the first light bulb
		false}

	// TODO: Add to SQLite db
	response.WriteEntity(l.Groups[thisLen]) // This might not follow spec!!!
}

func (l LightResource) getGroupAttr(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	response.WriteEntity(l.Groups[id])
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

	last_color := group.State.GetColor()
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
	SendState(&grouplight, last_color)

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
		"/api/{api_key}/groups",
		"Manage Groups",
		l._RegisterGroupsApi,
	)
}

func (l LightResource) _RegisterGroupsApi(ws *restful.WebService) {
	ws.Route(ws.GET("/").To(l.listGroups).
		Doc("list all groups").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("listGroups"))

	ws.Route(ws.POST("/").To(l.createGroup).
		Doc("create a group").
		Operation("createGroup"))

	ws.Route(ws.POST("{group-id}").To(l.getGroupAttr).
		Doc("get group attributes").
		Operation("getGroupAttr").
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")))

	ws.Route(ws.PUT("/{group-id}/action").To(l.updateGroupState).
		Doc("modify a groups's state").
		Operation("updateGroupState").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Reads(ColorState{}))
}
