package chromaticity

import (
	"fmt"
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
)

type ConfigInfo struct {
	Name         string               `json:"name"`
	Mac          string               `json:"mac"`
	Dhcp         bool                 `json:"dhcp"`
	Ipaddress    string               `json:"ipaddress"`
	Netmask      string               `json:"netmask"`
	Gateway      string               `json:"gateway"`
	Proxyaddress string               `json:"proxyaddress"`
	Proxyport    int                  `json:"proxyport"`
	UTC          string               `json:"UTC"`
	Whitelist    map[string]Developer `json:"whitelist"`
	Swversion    string               `json:"swversion"`
	//Apiversion     string               `json:"apiversion"` // 1.2.1 +
	Swupdate       UpdateInfo `json:"swupdate"`
	Linkbutton     bool       `json:"linkbutton"`
	Portalservices bool       `json:"portalservices"`
}

type Developer struct {
	LastUseDate string `json:"last use date"`
	CreateDate  string `json:"create date"`
	Name        string `json:"name"`
}

type UserInfo struct {
	DeviceType string `json:"devicetype"`
	UserName   string `json:"username"`
}

type UpdateInfo struct {
	Updatestate int    `json:"updatestate"`
	Url         string `json:"url"`
	Text        string `json:"text"`
	Notify      bool   `json:"notify"`
}

func (l LightResource) listInfo(request *restful.Request, response *restful.Response) {
	// Hack to make ip accurate, needed for some apps
	l.ConfigInfo.Ipaddress, _ = utils.GetHostPort(request.Request)
	l.ConfigInfo.Gateway = utils.DummyGateway(l.ConfigInfo.Ipaddress)
	response.WriteEntity(l)
}

func (l LightResource) listConfig(request *restful.Request, response *restful.Response) {
	// Hack to make ip accurate, needed for some apps
	l.ConfigInfo.Ipaddress, _ = utils.GetHostPort(request.Request)
	l.ConfigInfo.Gateway = utils.DummyGateway(l.ConfigInfo.Ipaddress)
	response.WriteEntity(l.ConfigInfo)
}

func (l LightResource) userCreate(request *restful.Request, response *restful.Response) {
	u := UserInfo{}
	request.ReadEntity(&u)
	if u.DeviceType == "" {
		response.WriteErrorString(400, "")
		return
	}
	if u.UserName == "" {
		u.UserName = "foobar"
	}
	l.ConfigInfo.Whitelist[u.UserName] = Developer{
		"2014-10-11T14:00:00",
		"2014-10-11T14:00:00",
		u.DeviceType,
	}
	fmt.Fprintf(response, `[{"success": {"username": "%s"}}]`, u.UserName)
}

func (l LightResource) _RegisterConfigApi(ws *restful.WebService) {
	ws.Route(ws.GET("{api_key}/config/all").To(l.listInfo).
		Doc("list all info").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("listInfo"))

	ws.Route(ws.GET("{api_key}/config/").To(l.listConfig).
		Doc("list all config info").
		Param(ws.PathParameter("api_key", "api key").DataType("string")).
		Operation("listConfig"))

	ws.Route(ws.POST("/").To(l.userCreate).
		Doc("create new api user").
		Operation("userCreate").
		Reads(UserInfo{}))
}

func (l LightResource) RegisterConfigApi(container *restful.Container) {
	utils.RegisterApis(
		container,
		"/api",
		"Manage Configuration",
		l._RegisterConfigApi,
	)
}

func NewConfigInfo() *ConfigInfo {
	return &ConfigInfo{
		"Philips hue",
		"b8:e8:56:29:15:98",
		true,
		"192.168.2.102",
		"255.255.255.0",
		"192.168.2.1",
		"",
		0,
		"2014-10-11T14:00:00",
		map[string]Developer{"newdev": Developer{
			"2014-10-11T14:00:00",
			"2014-10-11T14:00:00",
			"test user",
		},
		},
		"01003372",
		//"1.2.1",
		UpdateInfo{
			0,
			"",
			"",
			false,
		},
		true,
		false,
	}
}
