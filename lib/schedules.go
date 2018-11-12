package chromaticity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/evq/chromaticity/utils"
	"github.com/evq/go-restful"
)

var schedulesContainer *restful.Container

type Schedule struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Command     Command              `json:"command"`
	LocalTime   string               `json:"localtime"`
	Time        string               `json:"time"`
	Created     string               `json:"created"`
	Status      string               `json:"status"`
	Autodelete  bool                 `json:"autodelete"`
	Timer       *time.Timer          `json:"-"`
	Schedules   map[string]*Schedule `json:"-"`
	ID          string               `json:"-"`
}

type Command struct {
	Address string          `json:"address"`
	Body    json.RawMessage `json:"body"`
	Method  string          `json:"method"`
}

func (s *Schedule) execute() {
	err := s.executeOptionally(false)
	if err != nil {
		panic(err)
	}
}

func (s *Schedule) executeOptionally(test bool) error {
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
		log.Infof("[chromaticity/lib/schedules] Running schedule \"%s\"", s.ID)

		rec := httptest.NewRecorder()
		schedulesContainer.Dispatch(rec, req)
		resp := rec.Result()
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Error(err)
		}
		log.Debugf(
			"[chromaticity/lib/schedules] Ran schedule \"%s\": %s",
			s.ID,
			string(dump),
		)

		// reschedule
		t, err := utils.GetNextTimeFrom(s.LocalTime, nil)
		if err != nil {
			panic(err)
		}
		if t != nil {
			s.Timer = time.AfterFunc(time.Until(*t), s.execute)
		} else {
			// cleanup
			if s.Autodelete {
				delete(s.Schedules, s.ID)
			} else {
				s.Status = "disabled"
			}
		}
	}
	return nil
}

func (l LightResource) listSchedules(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l.Schedules)
}

func (l LightResource) deleteSchedule(request *restful.Request, response *restful.Response) {
	id := request.PathParameter("schedule-id")
	schedule := l.Schedules[id]
	if schedule == nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusNotFound, "404: Schedule could not be found.")
		return
	}
	schedule.Timer.Stop()
	delete(l.Schedules, id)
	WritePOSTSuccess(response, []string{fmt.Sprintf("/schedules/%s deleted.", id)})
}

func (l LightResource) createSchedule(request *restful.Request, response *restful.Response) {
	s := Schedule{Status: "enabled", Autodelete: true}
	err := request.ReadEntity(&s)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("400: Error in ReadEntity: %s", err.Error()))
		return
	}

	if len(s.LocalTime) == 0 && len(s.Time) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf(
			"400: Either time or localtime must be specified: %s", err.Error()))
		return
	}

	// hardcoded UTC as local tz
	if len(s.LocalTime) > 0 {
		s.Time = s.LocalTime
	}
	if len(s.Time) > 0 {
		s.LocalTime = s.Time
	}

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

	if len(s.Command.Body) == 0 || len(s.Command.Method) == 0 || len(s.Command.Address) == 0 {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("400: Command must be fully specified"))
		return
	}

	err = s.executeOptionally(true)
	if err != nil {
		response.AddHeader("Content-Type", "text/plain")
		response.WriteErrorString(http.StatusBadRequest, fmt.Sprintf("400: Error in executeOptionally: %s", err.Error()))
		return
	}

  log.Debugf(
    "[chromaticity/lib/schedules] Schedule \"%s\" next run at: %s",
    s.ID,
    t.String(),
  )

	s.Timer = time.AfterFunc(time.Until(*t), s.execute)
	s.Created = time.Now().Format(utils.DatetimeLayout)
	s.Schedules = l.Schedules

	max := 0
	for k, _ := range l.Schedules {
		e, err := strconv.Atoi(k)
		if err != nil {
			continue
		}
		if e > max {
			max = e
		}
	}
	s.ID = strconv.Itoa(max + 1)

	l.Schedules[s.ID] = &s

	WritePOSTSuccess(response, []string{s.ID})
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
	ws.Route(ws.DELETE("/{schedule-id}").To(l.deleteSchedule).
		Doc("delete schedule").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("deleteSchedule").
		Param(ws.PathParameter("schedule-id", "identifier of the schedule").DataType("int")).
		Writes([]SuccessResponse{}))
}
