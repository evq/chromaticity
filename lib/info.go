package chromaticity

import (
	"github.com/emicklei/go-restful"
	"github.com/evq/chromaticity/utils"
)

type ConfigInfo struct {
	Name           string               `json:"name"`
	Mac            string               `json:"mac"`
	Dhcp           bool                 `json:"dhcp"`
	Ipaddress      string               `json:"ipaddress"`
	Netmask        string               `json:"netmask"`
	Gateway        string               `json:"gateway"`
	Proxyaddress   string               `json:"proxyaddress"`
	Proxyport      int                  `json:"proxyport"`
	UTC            string               `json:"UTC"`
	Whitelist      map[string]Developer `json:"whitelist"`
	Swversion      string               `json:"swversion"`
	Swupdate       UpdateInfo           `json:"swupdate"`
	Linkbutton     bool                 `json:"linkbutton"`
	Portalservices bool                 `json:"portalservices"`
}

type Developer struct {
	LastUseDate string `json:"last use date"`
	CreateDate  string `json:"create date"`
	Name        string `json:"name"`
}

type UpdateInfo struct {
	Updatestate int    `json:"updatestate"`
	Url         string `json:"url"`
	Text        string `json:"text"`
	Notify      bool   `json:"notify"`
}

func (l LightResource) listInfo(request *restful.Request, response *restful.Response) {
	response.WriteEntity(l)
}

func (l LightResource) _RegisterConfigApi(ws *restful.WebService) {
	ws.Route(ws.GET("/all").To(l.listInfo).
		Doc("list all info").
		Operation("listInfo"))
}

func (l LightResource) RegisterConfigApi(container *restful.Container) {
	utils.RegisterApis(
		container,
		"/config",
		"Config api",
		l._RegisterConfigApi,
	)
}

func NewConfigInfo() *ConfigInfo {
	return &ConfigInfo{
		"Philips hue",
		"b8:e8:56:29:15:98",
		true,
		"192.168.1.185",
		"255.255.255.0",
		"192.168.1.1",
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
		UpdateInfo{
			0,
			"",
			"",
			false,
		},
		false,
		false,
	}
}
