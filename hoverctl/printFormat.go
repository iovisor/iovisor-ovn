package hoverctl

import (
	"fmt"
	"strings"
)

func LinkPrint(link LinkEntry) {
	fmt.Printf("link-id:%15s   from: %10s  to: %10s\n", link.Id, link.From, link.To)
}

func LinkListPrint(linkList map[string]LinkEntry) {
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

func PrintFirstNLines(s string, nlines int) {
	stringArray := strings.Split(s, "\n")
	for i := 0; i < nlines && i < len(stringArray); i++ {
		fmt.Printf("%s\n", stringArray[i])
	}
}
