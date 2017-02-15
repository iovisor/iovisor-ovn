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
package common

import (
	"os"

	"github.com/iovisor/iovisor-ovn/config"
	l "github.com/op/go-logging"
)

// var format = l.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)
var format = l.MustStringFormatter(`%{color}%{time:15:04:05} %{level:.4s} >%{color:reset} %{message}`)

//Logger initialization
func LogInit() {

	backend := l.NewLogBackend(os.Stderr, "", 0)

	if config.Debug == true {
		format = l.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)
	}

	backendFormatter := l.NewBackendFormatter(backend, format)

	backendLeveled := l.SetBackend(backendFormatter)

	backendLeveled.SetLevel(l.NOTICE, "")

	if config.Info == true {
		backendLeveled.SetLevel(l.INFO, "")
	}

	if config.Debug == true {
		backendLeveled.SetLevel(l.DEBUG, "")
	}

	l.SetBackend(backendLeveled)
}
