package chromaticity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"time"

	"github.com/evq/chromaticity/utils"
	"github.com/evq/go-restful"
)

var schedulesContainer *restful.Container

type Schedule struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Command     Command `json:"command"`
	LocalTime   string  `json:"localtime"`
	Time        string  `json:"time"`
	Create      string  `json:"create"`
	Status      string  `json:"status"`
	Autodelete  bool    `json:autdoelete"`
}

type Command struct {
	Address string          `json:"address"`
	Body    json.RawMessage `json:"body"`
	Method  string          `json:"method"`
}

func (s Schedule) execute() {
	err := s.executeOptionally(false)
	if err != nil {
		panic(err)
	}
}

func (s Schedule) executeOptionally(test bool) error {
	// execute command

	b, err := s.Command.Body.MarshalJSON()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(s.Command.Method, s.Command.Address, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	if !test {
		rec := httptest.NewRecorder()
		schedulesContainer.Dispatch(rec, req)
		resp := rec.Result()
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(string(dump))

		// reschedule
		t, err := utils.GetNextTimeFrom(s.LocalTime, nil)
		if err != nil {
			panic(err)
		}
		if t != nil {
			// if valid, schedule again
			time.AfterFunc(time.Until(*t), s.execute)
		}
	}
	return nil
}

func (l LightResource) listSchedules(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Schedules)
}

func (l LightResource) createSchedule(request *restful.Request, response *restful.Response) {
	s := Schedule{}
	request.ReadEntity(&s)
	t, err := utils.GetNextTimeFrom(s.LocalTime, nil)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("400: Error in GetNextTimeFrom: %s", err.Error()))
		return
	}
	if t == nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("400: No time returned from GetNextTimeFrom"))
		return
	}
	err = s.executeOptionally(true)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("400: Error in executeOptionally: %s", err.Error()))
		return
	}

	time.AfterFunc(time.Until(*t), s.execute)

	response.WriteEntity(l.Schedules)
}

func (l LightResource) RegisterSchedulesApi(container *restful.Container) {
	utils.RegisterApis(
		container,
		"/api/{api_key}/schedules",
		"Manage Schedules",
		l._RegisterSchedulesApi,
	)
	schedulesContainer = container
}

func (l LightResource) _RegisterSchedulesApi(ws *restful.WebService) {
	ws.Route(ws.GET("/").To(l.listSchedules).
		Doc("list all schedules").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("listSchedules"))
	ws.Route(ws.POST("/").To(l.createSchedule).
		Doc("create schedule").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("createSchedule").
		Reads(Schedule{}).
		Writes([]SuccessResponse{}))

	/*
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
	*/
}
