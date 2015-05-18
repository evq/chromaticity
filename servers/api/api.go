package api

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/evq/go-restful"
	"github.com/evq/go-restful/swagger"
	"github.com/evq/chromaticity/backends"
	"github.com/evq/chromaticity/static"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/chromaticity/utils"
	"net/http"
	"mime"
	"strings"
	"text/template"
)

type Service struct {
	IP   string
	Port string
}

type AuthHandler struct {
	chainedHandler http.Handler
}

type AssetHandler struct { }

func (a *AssetHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	log.Debug(req.URL.Path)
	if strings.HasSuffix(req.URL.Path, "/") {
		req.URL.Path = req.URL.Path + "index.html"
	}
	data, err := static.Asset("static/" + req.URL.Path[1:])
	if err == nil {
		split := strings.Split(req.URL.Path, ".")
		ext := "." + split[len(split)-1]
		ct := mime.TypeByExtension(ext)
		resp.Header().Set("Content-Type", ct)
		resp.Write(data)
	} else {
		http.Error(resp, "File not found", 404)
	}
}

func SsdpDescription(resp http.ResponseWriter, req *http.Request) {
	t := template.New("description.xml")
	data, err := static.Asset("static/description.xml")
	t, err = t.Parse(string(data))
	if err != nil {
		log.Error(err)
	}
	s := Service{}
	s.IP, s.Port = utils.GetHostPort(req)
	t.Execute(resp, s)
}

func ReqRewriter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	contenttype := req.Request.Header.Get("Content-Type")
	if contenttype != "application/json" {
		log.Info(fmt.Sprintf(
			"[chromaticity/servers/api] Rewriting Content-Type: %s -> application/json",
			contenttype,
		))
		// Some hue apps don't set content type correctly AFAIK
		req.Request.Header.Set("Content-Type", "application/json")
	}

	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	resp.Header().Set("Access-Control-Allow-Methods", "HEAD,GET,PUT,DELETE,OPTIONS")

	chain.ProcessFilter(req, resp)
}

func ReqLogger(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	chain.ProcessFilter(req, resp)

	api_key := req.PathParameter("api_key")

	log.Info(fmt.Sprintf(
		"[chromaticity/servers/api] %s %s %s %s %s %d",
		strings.Split(req.Request.RemoteAddr, ":")[0],
		api_key,
		req.Request.Method,
		req.Request.URL.RequestURI(),
		req.Request.Header.Get("Content-Type"),
		resp.StatusCode(),
	))

	var temp interface{}
	err := req.ReadEntity(&temp)
	if err != nil {
		return
	}
	content, err := json.Marshal(temp)
	if err != nil {
		return
	}
	log.Debug("[chromaticity/servers/api] " + string(content))
}

func StartServer(port string, configfile string) {
	l := &chromaticity.LightResource{}
	l.ConfigInfo = chromaticity.NewConfigInfo()
	l.Schedules = map[string]string{}
	backends.Load(l, configfile)

	restful.SetLogger(log.StandardLogger())

	wsContainer := restful.NewContainer()
	wsContainer.Filter(ReqLogger)
	wsContainer.Filter(ReqRewriter)

	// Register apis
	l.RegisterConfigApi(wsContainer)
	l.RegisterLightsApi(wsContainer)
	l.RegisterGroupsApi(wsContainer)
	backends.RegisterDiscoveryApi(wsContainer, l)

	// Start goroutines to send pixel data
	backends.Sync()

	// Uncomment to add some swagger
	config := swagger.Config{
		WebServices:     wsContainer.RegisteredWebServices(),
		WebServicesUrl:  "/",
		ApiPath:         "/swagger/apidocs.json",
		SwaggerPath:     "/swagger/apidocs/",
	}

	//Container just for swagger
	swContainer := restful.NewContainer()
	swagger.RegisterSwaggerService(config, swContainer)
	http.Handle("/swagger/", swContainer)
	http.Handle("/apidocs/", &AssetHandler{})

	http.HandleFunc("/description.xml", SsdpDescription)

	http.Handle("/api", wsContainer)
	http.Handle("/api/", wsContainer)

	log.Info("[chromaticity/servers/api] start listening on localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
