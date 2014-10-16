package api

import (
	"github.com/emicklei/go-restful"
	"github.com/emicklei/go-restful/swagger"
	"github.com/evq/chromaticity/backends"
	chromaticity "github.com/evq/chromaticity/lib"
	"log"
	"net/http"
	"os"
	"strings"
  "encoding/json"
  "fmt"
  //"io/ioutil"
)

type AuthHandler struct {
	chainedHandler http.Handler
}

type UserInfo struct {
  DeviceType string `json:"devicetype"`
  UserName
}

type UserName struct {
  UN string `json:"username"`
}

// Whitelist goes here

func UserCreate(resp http.ResponseWriter, req *http.Request) {
    decoder := json.NewDecoder(req.Body)
    var u UserInfo
    err := decoder.Decode(&u)
    if err != nil {
      // FIXME
      fmt.Fprintf(resp, "ERROR")
    }
    jsonUN, _ := json.Marshal(u.UserName)
    fmt.Fprintf(resp, `[{"success": %s}]`, string(jsonUN))
}

func (a *AuthHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
  if req.Header.Get("Content-Type") == "" {
    req.Header.Set("Content-Type","application/json")
  }
  resp.Header().Set("Access-Control-Allow-Origin", "*")
  resp.Header().Set("Access-Control-Allow-Headers", "Content-Type")
  resp.Header().Set("Access-Control-Allow-Methods", "HEAD,GET,PUT,DELETE,OPTIONS")

  url := req.URL.Path
  log.Printf("%s %s\n", req.Method, url)

  if req.Method == "OPTIONS" {
    return
  }

  if url == "/api/" {
    UserCreate(resp, req)
    return
  }

	url = url[len("/api/"):]
	i := strings.Index(url, "/")

  if i == -1 {
    req.URL.Path = "/config/all"
    a.chainedHandler.ServeHTTP(resp, req)
    return
  }

	token := url[:i]

  // Strip token for downstream handler
	req.URL.Path = url[len(token):]

	// Do token matching here :)

	a.chainedHandler.ServeHTTP(resp, req)
}

func StartApiServer() {
	l := &chromaticity.LightResource{}
  l.ConfigInfo = chromaticity.NewConfigInfo()
	backends.Load(l)

	wsContainer := restful.NewContainer()

	// Register apis
  l.RegisterConfigApi(wsContainer)
	l.RegisterLightsApi(wsContainer)
	l.RegisterGroupsApi(wsContainer)
	backends.RegisterDiscoveryApi(wsContainer, l)

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

	http.HandleFunc("/api", UserCreate)
	http.Handle("/api/", &AuthHandler{wsContainer})

	log.Printf("[chromaticity] start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
