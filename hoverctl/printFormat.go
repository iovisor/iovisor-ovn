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
	"fmt"
	"strings"
)

func LinkPrint(link Link) {
	fmt.Printf("link-id:%15s   from: %10s (%d) to: %10s (%d)\n", link.Id, link.From, link.FromId, link.To, link.ToId)
}

func LinkListPrint(linkList map[string]Link) {
	for _, link := range linkList {
		LinkPrint(link)
	}
}

func ModulePrint(module Module) {
	fmt.Printf("module-id:%8s	display_name:%8s	module_type:%10s\n", module.Id, module.DisplayName, module.ModuleType)
	// if module.Config["code"] != nil {
	// 	PrintFirstNLines(module.Config["code"].(string), 10)
	// 	fmt.Printf("[...]\n\n")
	// }
}

func ModuleListPrint(moduleList map[string]Module) {
	for _, module := range moduleList {
		ModulePrint(module)
	}
}

func ExternalInterfacePrint(externalInterface ExternalInterface) {
	fmt.Printf("interface-id:%15s   id: %5s\n", externalInterface.Name, externalInterface.Id)
}

func ExternalInterfacesListPrint(externalInterfacesList map[string]ExternalInterface) {
	for _, externalInterface := range externalInterfacesList {
		ExternalInterfacePrint(externalInterface)
	}
}

func TablePrint(table map[string]TableEntry) {
	for _, tableEntry := range table {
		TableEntryPrint(tableEntry)
	}
}

func TableEntryPrint(tableEntry TableEntry) {
	fmt.Printf("key: %15s	value: %15s\n", tableEntry.Key, tableEntry.Value)
}

func PrintFirstNLines(s string, nlines int) {
	stringArray := strings.Split(s, "\n")
	for i := 0; i < nlines && i < len(stringArray); i++ {
		fmt.Printf("%s\n", stringArray[i])
	}
}
