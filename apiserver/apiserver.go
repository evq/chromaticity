package apiserver

import (
	"github.com/emicklei/go-restful"
	//"github.com/emicklei/go-restful/swagger"
	"github.com/evq/chromaticity/backends"
	chromaticity "github.com/evq/chromaticity/lib"
	"log"
	"net/http"
)

func StartApiServer() {
	l := &chromaticity.LightResource{}
	backends.Load(l)

	wsContainer := restful.NewContainer()

	// Register apis
	l.RegisterLightsApi(wsContainer)
	l.RegisterGroupsApi(wsContainer)
	backends.RegisterDiscoveryApi(wsContainer, l)

	// Uncomment to add some swagger
  //config := swagger.Config{
    //WebServices:     wsContainer.RegisteredWebServices(), 
    //WebServicesUrl:  "http://localhost:8080/api/foo",
    //ApiPath:         "/swagger/apidocs.json",
    //SwaggerPath:     "/swagger/apidocs/",
    //SwaggerFilePath: "/Users/ev/Documents/syncd/projects/swagger-ui/dist"}

	// Container just for swagger
	//swContainer := restful.NewContainer()
	//swagger.RegisterSwaggerService(config, swContainer)
	//http.Handle("/swagger/", swContainer)

	// FIXME foo needs to be replaced with authentication :)
	http.Handle("/api/foo/", http.StripPrefix("/api/foo", wsContainer))

	log.Printf("[chromaticity] start listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
