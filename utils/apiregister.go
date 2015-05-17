package utils

import (
	"github.com/evq/go-restful"
)

// Helper method to allow registering apis on the same endpoint
func RegisterApis(container *restful.Container, rootPath string, docString string, apis func(*restful.WebService)) {
	newService := true
	ws := new(restful.WebService)

	webservices := container.RegisteredWebServices()
	for i := range webservices {
		if webservices[i].RootPath() == rootPath {
			newService = false
			ws = webservices[i]
		}
	}

	if newService {
		ws.Path(rootPath).
			Doc(docString).
			Consumes("*/*").
			Produces(restful.MIME_JSON) // you can specify this per route as well
	}

	apis(ws)

	if newService {
		container.Add(ws)
	}
}
