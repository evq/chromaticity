package api

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/evq/chromaticity/backends"
	"github.com/evq/chromaticity/utils"
	chromaticity "github.com/evq/chromaticity/lib"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"fmt"
)

type Service struct {
	IP   string
	Port string
}

type AuthHandler struct {
	chainedHandler http.Handler
}

func SsdpDescription(resp http.ResponseWriter, req *http.Request) {
	t := template.New("description.xml")
	var err error
	t, err = t.ParseFiles(os.Getenv("GOPATH") + "/src/github.com/evq/chromaticity/servers/ssdp/description.xml")
	if err != nil {
		log.Println(err)
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

	//log.Printf(req.Header["Connection"][0])

	//req.Header["Connection"] = []string{"keep-alive"}


	url := req.URL.Path
	log.Printf("%s %s\n", req.Method, url)

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
	fmt.Println(req.URL.Path)

	// Do token matching here :)

	a.chainedHandler.ServeHTTP(resp, req)
}

func StartServer(port string) {
	l := &chromaticity.LightResource{}
	l.ConfigInfo = chromaticity.NewConfigInfo()
	l.Schedules = map[string]string{}
	backends.Load(l)

	wsContainer := restful.NewContainer()

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
		WebServicesUrl:  "http://localhost/api/swagger",
		ApiPath:         "/swagger/apidocs.json",
		SwaggerPath:     "/swagger/apidocs/",
		SwaggerFilePath: os.Getenv("GOPATH") + "/src/github.com/evq/chromaticity/swagger-ui/dist",
	}

	//Container just for swagger
	swContainer := restful.NewContainer()
	swagger.RegisterSwaggerService(config, swContainer)
	http.Handle("/swagger/", swContainer)

	http.HandleFunc("/description.xml", SsdpDescription)

	http.Handle("/api/", &AuthHandler{wsContainer})


	log.Printf("[chromaticity] start listening on localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
