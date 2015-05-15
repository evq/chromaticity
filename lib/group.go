package chromaticity

import (
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
	"net/http"
	"encoding/json"
	"strings"
	"io"
	"log"
	"fmt"
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
	GroupType string  `json:"type"`
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
		thisState := & ((* l.Lights[lights[0]]).GetState().ColorState)
		ginfo := GroupInfo{ lights,
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
	var group Group
	if id == "0" {
		ids := []string{}
		for k, _ := range l.Lights {
			ids = append(ids, k)
		}
		group = Group{
			GroupInfo{ids, "All Lights", "0"},
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
		cs = (*l.Lights[id]).GetState().ColorState
		cs.Xy = []float64{cs.Xy[0], cs.Xy[1]}
		request.ReadEntity(&cs)
		UpdateColorState(l.Lights[id], cs)
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
		Param(ws.PathParameter("group-id", "identifier of the group").DataType("int")).
		Reads(ColorState{}))
}
