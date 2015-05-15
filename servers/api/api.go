package api

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/evq/chromaticity/backends"
	"github.com/evq/chromaticity/static"
	chromaticity "github.com/evq/chromaticity/lib"
	"github.com/evq/chromaticity/utils"
	"net/http"
	urllib "net/url"
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
func (a *AuthHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	// Always assume we are getting json
	req.Header.Set("Content-Type", "application/json")

	resp.Header().Set("Access-Control-Allow-Origin", "*")
	resp.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	resp.Header().Set("Access-Control-Allow-Methods", "HEAD,GET,PUT,DELETE,OPTIONS")

	url := req.URL.Path

	if req.Method == "OPTIONS" {
		return
	}

	if len(url) < len("/api") {
		return
	}

	url = url[len("/api"):]
	if url == "" || url == "/" {
		req.URL.Path = "/"
		a.chainedHandler.ServeHTTP(resp, req)
		return
	}
	url = url[1:]

	i := strings.Index(url, "/")

	if i == -1 {
		req.URL.Path = "/config/all"
		a.chainedHandler.ServeHTTP(resp, req)
		return
	}

	token := url[:i]

	// Strip token for downstream handler
	req.URL.Path = url[len(token):]

	req.URL.User = urllib.User(token)
	// Do token matching here :)

	a.chainedHandler.ServeHTTP(resp, req)
}

func ReqLogger(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	chain.ProcessFilter(req, resp)

	username := ""
	user := req.Request.URL.User
	if user != nil {
		username = user.Username()
	}

	log.Info(fmt.Sprintf(
		"[chromaticity/servers/api] %s %s %s %s %s %d",
		strings.Split(req.Request.RemoteAddr, ":")[0],
		username,
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

func StartServer(port string) {
	l := &chromaticity.LightResource{}
	l.ConfigInfo = chromaticity.NewConfigInfo()
	l.Schedules = map[string]string{}
	backends.Load(l)

	restful.SetLogger(log.StandardLogger())

	wsContainer := restful.NewContainer()
	wsContainer.Filter(ReqLogger)

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

	//http.Handle("/api/", &AuthHandler{wsContainer})
	http.Handle("/api/", wsContainer)

	log.Info("[chromaticity/servers/api] start listening on localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
