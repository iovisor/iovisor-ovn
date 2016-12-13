// Copyright 2016 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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

type Module struct {
	Id          string                 `json:"id"`
	ModuleType  string                 `json:"module_type"`
	DisplayName string                 `json:"display_name"`
	Perm        string                 `json:"permissions"`
	Config      map[string]interface{} `json:"config"`
}

type ExternalInterface struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Link struct {
	Id     string `json:"id"`
	From   string `json:"from"`
	To     string `json:"to"`
	FromId int    `json:"from-id"`
	ToId   int    `json:"to-id"`
}

type TableEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
