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
)

type AuthHandler struct {
	chainedHandler http.Handler
}

func (a *AuthHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	url := req.URL.Path[len("/api/"):]
	i := strings.Index(url, "/")
	token := url[:i]

  // Strip token for downstream handler
	req.URL.Path = url[len(token):]

	// Do token matching here :)

	a.chainedHandler.ServeHTTP(resp, req)
}

func StartApiServer() {
	l := &chromaticity.LightResource{}
	backends.Load(l)

	wsContainer := restful.NewContainer()

	// Register apis
	l.RegisterLightsApi(wsContainer)
	l.RegisterGroupsApi(wsContainer)
	backends.RegisterDiscoveryApi(wsContainer, l)

	// Uncomment to add some swagger
	config := swagger.Config{
		WebServices:     wsContainer.RegisteredWebServices(),
		WebServicesUrl:  "http://localhost:8080/api/swagger",
		ApiPath:         "/swagger/apidocs.json",
		SwaggerPath:     "/swagger/apidocs/",
		SwaggerFilePath: os.Getenv("GOPATH") + "/src/github.com/evq/chromaticity/swagger-ui/dist",
  }

	//Container just for swagger
	swContainer := restful.NewContainer()
	swagger.RegisterSwaggerService(config, swContainer)
	http.Handle("/swagger/", swContainer)

	http.Handle("/api/", &AuthHandler{wsContainer})

	log.Printf("[chromaticity] start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
