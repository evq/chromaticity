package chromaticity

import (
	"github.com/evq/go-restful"
	"text/template"
	"bytes"
)

type ErrorResponse struct {
	Type int `json:"type"`
	Address string `json:"address"`
	Description string `json:"description"`
}

// Error Types
const (
	JSONError = 2
	ResourceError = 3
	ValueError = 7
	ROGroupError = 305
)

var ErrorDescriptions map[int]string = map[int]string{
	JSONError: "body contains invalid JSON",
	ResourceError: "resource, {{.resource}}, not available",
	ValueError: "invalid value, {{.value}}, for parameter, {{.parameter}}",
	ROGroupError: "It is not allowed to update or delete group of this type",
}

var ErrorStatusCodes map[int]int = map[int]int{
	JSONError: 400,
	ResourceError: 404,
	ValueError: 400,
	ROGroupError: 403,
}

type SuccessResponse struct {
	 Field interface{} `json:"success"`
}

func WriteError(response *restful.Response, errorType int, addr string, vars *map[string]string) {
	descr := ErrorDescriptions[errorType]
	if vars != nil {
		var b bytes.Buffer
		template, err := template.New("description").Parse(descr)
		if err != nil { panic(err) }
		template.Execute(&b, *vars)
		descr = b.String()
	}
	response.WriteHeader(ErrorStatusCodes[errorType])
	response.WriteEntity(ErrorResponse{
		errorType,
		addr,
		descr,
	})
}

func WritePOSTSuccess(response *restful.Response, ids []string) {
	resp := make([]SuccessResponse, len(ids), len(ids))
	for i := range ids {
		resp[i].Field = map[string]string{"id":ids[i]}
	}
	response.WriteEntity(resp)
}

func WritePUTSuccess(response *restful.Response, fields map[string]interface{}) {
	resp := make([]SuccessResponse, len(fields), len(fields))
	i := 0
	for k, v := range fields {
		resp[i].Field = map[string]interface{}{k:v}
		i++
	}
	response.WriteEntity(resp)
}

// Unfortunately the Hue API seems to be inconsistant in response format
func WritePUTSuccessExplicit(response *restful.Response, fields map[string]interface{}) {
	resp := make([]SuccessResponse, len(fields), len(fields))
	i := 0
	for k, v := range fields {
		resp[i].Field = map[string]interface{}{
			"address": k,
			"value": v,
		}
		i++
	}
	response.WriteEntity(resp)
}

func WriteDELETESuccess(response *restful.Response, ids []string) {
	resp := make([]SuccessResponse, len(ids), len(ids))
	for i := range ids {
		resp[i].Field = ids[i] + " deleted."
	}
	response.WriteEntity(resp)
}

