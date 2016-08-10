// Copyright 2016 PLUMgrid
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

package politoctrl

import (
	_ "io/ioutil"
	"log"
	"os"
)

var (
	Debug *log.Logger
	Info  *log.Logger
	Warn  *log.Logger
	Error *log.Logger
)

func logInit() {
	Debug = log.New(os.Stdout, "DBG: ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "INF: ", log.Ldate|log.Ltime)
	Warn = log.New(os.Stdout, "WRN: ", log.Ldate|log.Ltime)
	Error = log.New(os.Stderr, "ERR: ", log.Ldate|log.Ltime)
}

func init() {
	logInit()
}
