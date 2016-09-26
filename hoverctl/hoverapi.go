package hoverctl

import (
	_ "net"
	"net/http"
)

type Dataplane struct {
	client  *http.Client
	baseUrl string
	id      string
}

func NewDataplane() *Dataplane {
	client := &http.Client{}
	d := &Dataplane{
		client: client,
	}

	return d
}

func (d *Dataplane) Init(baseUrl string) error {
	d.baseUrl = baseUrl
	return nil
}

type ModuleEntry struct {
	Id          string                 `json:"id"`
	ModuleType  string                 `json:"module_type"`
	DisplayName string                 `json:"display_name"`
	Perm        string                 `json:"permissions"`
	Config      map[string]interface{} `json:"config"`
}

type ModuleList struct {
	ListModules []ModuleEntry
}

type ExternalInterface struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type LinkEntry struct {
	Id   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
}

type ModuleTableEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
