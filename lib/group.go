package chromaticity

import (
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
	"net/http"
)

type Group struct {
	GroupInfo
	State    *ColorState `json:"action"`
	ReadOnly bool        `json:"-"`
	//Scenes
}

type GroupInfo struct {
	LightIDs []string `json:"lights"`
	Name     string   `json:"name"`
}

func (l LightResource) listGroups(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Groups)
}

func (l LightResource) updateGroupState(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("group-id")
	// Special case, hidden group of all lights
	var group Group
	if id == "0" {
		ids := []string{}
		for k, _ := range l.Lights {
			ids = append(ids, k)
		}
		group = Group{
			GroupInfo{ids, "All Lights"},
			&NewState().ColorState,
			true,
		}
	} else {
		group = l.Groups[id]
	}
	if len(group.Name) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Group could not be found.")
		return
	}

	cs := *group.State
	cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
	request.ReadEntity(&cs)
	_UpdateColorState(group.State, cs)

	for i := range group.LightIDs {
		id := group.LightIDs[i]
		//cs = (*l.Lights[id]).GetState().ColorState
		//cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
		//request.ReadEntity(&cs)
		//UpdateColorState(l.Lights[id], cs)
		UpdateColorState((*l.Lights[id]).GetState(), request)
		SendState(l.Lights[id])
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
