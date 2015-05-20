package backends

import (
	"encoding/json"
	"github.com/evq/go-restful"
	"github.com/evq/chromaticity/backends/kinetclient"
	//"github.com/evq/chromaticity/backends/limitlessclient"
	"github.com/evq/chromaticity/backends/opclient"
	"github.com/evq/chromaticity/backends/zigbeeclient"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/chromaticity/utils"
	"io/ioutil"
)

var allBackends = []Backend{
	&kinetclient.Backend{},
	&opclient.Backend{},
	//&limitlessclient.Backend{},
	&zigbeeclient.Backend{},
}

type Backend interface {
	//ExportLights(l *chromaticity.LightResource) (string)
	ImportLights(l *chromaticity.LightResource, from []byte)
	DiscoverLights(l *chromaticity.LightResource)
	GetType() string
	Sync()
}

type DiscoverResource struct {
	*chromaticity.LightResource
}

func RegisterDiscoveryApi(container *restful.Container, l *chromaticity.LightResource) {
	d := DiscoverResource{l}
	utils.RegisterApis(
		container,
		"/api/{api_key}/lights",
		"Manage Lights",
		d._RegisterDiscoveryApi,
	)
}

func (d DiscoverResource) _RegisterDiscoveryApi(ws *restful.WebService) {
	ws.Route(ws.POST("/").To(d.searchLights).
		Doc("search for lights").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("searchLights").
	  Writes([]chromaticity.SuccessResponse{}))
}

func (d DiscoverResource) searchLights(request *restful.Request, response *restful.Response) {
	Discover(d.LightResource)
	chromaticity.WritePUTSuccess(response, map[string]interface{}{
		"/lights": "Searching for new devices",
	})
}

func Sync() {
	for i := range allBackends {
		allBackends[i].Sync()
	}
}

func Load(l *chromaticity.LightResource, configfile string) {
	if l.Groups == nil {
		l.Groups = map[string]*chromaticity.Group{}
	}
	if l.Lights == nil {
		l.Lights = map[string]*chromaticity.Light{}
	}

	data, _ := ioutil.ReadFile(configfile)

	var exportData map[string]interface{}
	json.Unmarshal(data, &exportData)

	for i := range allBackends {
		backend := allBackends[i]
		str, _ := json.Marshal(exportData[backend.GetType()])
		backend.ImportLights(l, str)
	}
}

func Discover(l *chromaticity.LightResource) {
	if l.Groups == nil {
		l.Groups = map[string]*chromaticity.Group{}
	}
	if l.Lights == nil {
		l.Lights = map[string]*chromaticity.Light{}
	}
	for i := range allBackends {
		allBackends[i].DiscoverLights(l)
	}
}
