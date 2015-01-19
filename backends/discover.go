package backends

import (
	"encoding/json"
	"github.com/emicklei/go-restful"
  "github.com/evq/chromaticity/backends/kinetclient"
  "github.com/evq/chromaticity/backends/opclient"
  "github.com/evq/chromaticity/backends/limitlessclient"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/chromaticity/utils"
	"io/ioutil"
	"os/user"
)

var allBackends = []Backend{
  kinetclient.Backend{},
  &opclient.Backend{},
  &limitlessclient.Backend{},
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

const DATA_FILE = "/.chromaticity/data.json"

func RegisterDiscoveryApi(container *restful.Container, l *chromaticity.LightResource) {
	d := DiscoverResource{l}
	utils.RegisterApis(
		container,
		"/lights",
		"Manage Lights",
		d._RegisterDiscoveryApi,
	)
}

func (d DiscoverResource) _RegisterDiscoveryApi(ws *restful.WebService) {
	ws.Route(ws.POST("/").To(d.searchLights).
		Doc("search for lights").
		Operation("searchLights"))
}

func (d DiscoverResource) searchLights(request *restful.Request, response *restful.Response) {
	Discover(d.LightResource)
}

func Sync() {
  for i := range allBackends {
    allBackends[i].Sync()
  }
}

func Load(l *chromaticity.LightResource) {
	if l.Groups == nil {
		l.Groups = map[string]chromaticity.Group{}
	}
	if l.Lights == nil {
		l.Lights = map[string]*chromaticity.Light{}
	}

	// FIXME Errors :)
	usr, _ := user.Current()
	//if err != nil {
	//log.Fatal( err )
	//}
	data, _ := ioutil.ReadFile(usr.HomeDir + DATA_FILE)

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
		l.Groups = map[string]chromaticity.Group{}
	}
	if l.Lights == nil {
		l.Lights = map[string]*chromaticity.Light{}
	}
	for i := range allBackends {
		allBackends[i].DiscoverLights(l)
	}
}
