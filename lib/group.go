package chromaticity

import (
	log "github.com/Sirupsen/logrus"
	"github.com/evq/chromaticity/utils"
	"github.com/evq/go-restful"
	"github.com/lucasb-eyer/go-colorful"
	"strconv"
	"reflect"
)

const Luminaire string = "Luminaire"
const Lightsource string = "Lightsource"
const LightGroup string = "LightGroup"

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
	for l := range g.Lights {
		if (*g.Lights[l]).GetType() == "zigbee" {
			return "zigbee"
		}
	}
	return "group"
}

func (l LightResource) listGroups(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Groups)
}

func (l LightResource) createGroup(request *restful.Request, response *restful.Response) {
	ginfo := GroupInfo{}
	err := request.ReadEntity(&ginfo)
	if err != nil { WriteError(response, JSONError, "/groups", nil); return }

	thisLen := strconv.Itoa(len(l.Groups) + 1)

	// Copy state from first bulb
	thisState := (*l.Lights[ginfo.LightIDs[0]]).GetState().ColorState

	theseLights := make([]*Light, len(ginfo.LightIDs), len(ginfo.LightIDs))

	for i := range ginfo.LightIDs { theseLights[i] = l.Lights[ginfo.LightIDs[i]] }

	ginfo.GroupType = LightGroup

	l.Groups[thisLen] = &Group{
		ginfo,
		thisState,
		false,
		theseLights,
	}

	WritePOSTSuccess(response, []string{"/groups/" + thisLen})
}

func (l LightResource) deleteGroup(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	if nid, err := strconv.Atoi(id); err != nil || nid > len(l.Groups) {
		WriteError(response, ResourceError, "/groups/" + id, &map[string]string{"resource":"/groups/" + id})
		return
	}
	if id == "0" { WriteError(response, ROGroupError, "/groups/" + id, nil); return }
	group := l.Groups[id]
	if group.GroupType == Luminaire || group.GroupType == Lightsource {
		WriteError(response, ROGroupError, "/groups/" + id, nil)
		return
	}

	if group.State.EffectRoutine != nil && !group.State.EffectRoutine.Done {
		for i := range group.Lights {
			state := (*group.Lights[i]).GetState()
			if state.EffectRoutine != nil {
				log.Debug("[chromaticity/lib/group] Canceling effect")
				if !state.EffectRoutine.Done {
					state.EffectRoutine.Done = true
					state.EffectRoutine.Signal <- true
					close(state.EffectRoutine.Signal)
				}
				state.EffectRoutine = nil
				log.Debug("[chromaticity/lib/group] Done canceling effect")
			}
		}
	}

	delete(l.Groups, id)

	WriteDELETESuccess(response, []string{"/groups/" + id})
}

func (l LightResource) getGroup(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	if nid, err := strconv.Atoi(id); err != nil || nid > len(l.Groups) {
		WriteError(response, ResourceError, "/groups/" + id, &map[string]string{"resource":"/groups/" + id})
		return
	}
	response.WriteEntity(l.Groups[id])
}

func (l LightResource) updateGroup(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	if nid, err := strconv.Atoi(id); err != nil || nid > len(l.Groups) {
		WriteError(response, ResourceError, "/groups/" + id, &map[string]string{"resource":"/groups/" + id})
		return
	}
	if id == "0" { WriteError(response, ROGroupError, "/groups/" + id, nil); return }

	group := l.Groups[id]
	if group.GroupType == Luminaire || group.GroupType == Lightsource {
		WriteError(response, ROGroupError, "/groups/" + id, nil)
		return
	}

	updated := GroupInfo{}
	err := request.ReadEntity(&updated)
	if err != nil { WriteError(response, JSONError, "/groups/" + id, nil); return }
	if updated.Name != group.Name {
		group.Name = updated.Name
	}
	if updated.LightIDs != nil && !reflect.DeepEqual(updated.LightIDs, group.LightIDs) {
		group.LightIDs = updated.LightIDs
		group.Lights = make([]*Light, len(group.LightIDs), len(group.LightIDs))
		for i := range group.LightIDs { group.Lights[i] = l.Lights[group.LightIDs[i]] }
	}

	reqmap := map[string]interface{}{}
	request.ReadEntity(&reqmap)
	respmaps := make(map[string]interface{})
	for k, v := range reqmap { respmaps["/groups/" + id + "/" + k] = v }

	WritePUTSuccess(response, respmaps)
}

func (l LightResource) updateGroupState(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	if nid, err := strconv.Atoi(id); err != nil || nid > len(l.Groups) {
		WriteError(response, ResourceError, "/groups/" + id, &map[string]string{"resource":"/groups/" + id})
		return
	}
	// Special case, hidden group of all lights
	// FIXME doesn't work for effects
	var group *Group
	if id == "0" {
		ids := []string{}
		for k, _ := range l.Lights {
			ids = append(ids, k)
		}
		group = NewGroup(
			&l,
			GroupInfo{ids, "All Lights", "0"},
			NewState().ColorState,
			true,
		)
	} else {
		group = l.Groups[id]
	}

	last_color := group.State.GetColor()
	UpdateColorState(group.State, request)

	for i := range group.LightIDs {
		id := group.LightIDs[i]
		state := (*l.Lights[id]).GetState()
		UpdateColorState(state.ColorState, request)
		if state.EffectRoutine != nil {
			log.Debug("[chromaticity/lib/group] Canceling effect")
			if !state.EffectRoutine.Done {
				state.EffectRoutine.Done = true
				state.EffectRoutine.Signal <- true
				close(state.EffectRoutine.Signal)
			}
			state.EffectRoutine = nil
			log.Debug("[chromaticity/lib/group] Done canceling effect")
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

	reqmap := map[string]interface{}{}
	request.ReadEntity(&reqmap)
	respmaps := make(map[string]interface{})
	for k, v := range reqmap { respmaps["/groups/" + id + "/action/" + k] = v }

	WritePUTSuccessExplicit(response, respmaps)
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
		Operation("listGroups").
		Writes([]Group{}))

	ws.Route(ws.POST("/").To(l.createGroup).
		Doc("create a group").
		Operation("createGroup").
		Reads(GroupInfo{}).
		Writes([]SuccessResponse{}))

	ws.Route(ws.GET("{group-id}").To(l.getGroup).
		Doc("get group attributes").
		Operation("getGroup").
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Writes(Group{}))

	ws.Route(ws.PUT("{group-id}").To(l.updateGroup).
		Doc("update group attributes").
		Operation("updateGroup").
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Reads(GroupInfo{}).
		Writes([]SuccessResponse{}))

	ws.Route(ws.DELETE("{group-id}").To(l.deleteGroup).
		Doc("delete a group").
		Operation("deleteGroup").
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Writes([]SuccessResponse{}))

	ws.Route(ws.PUT("/{group-id}/action").To(l.updateGroupState).
		Doc("modify a groups's state").
		Operation("updateGroupState").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Reads(ColorState{}).
		Writes([]SuccessResponse{}))
}
