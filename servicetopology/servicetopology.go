// Copyright 2017 Politecnico di Torino
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
package servicetopology

import (
	"errors"
	"fmt"

	"github.com/mvbpolito/gosexy/to"
	"github.com/mvbpolito/gosexy/yaml"

	"github.com/iovisor/iovisor-ovn/config"
	"github.com/iovisor/iovisor-ovn/hover"

	"github.com/iovisor/iovisor-ovn/iomodules"
	"github.com/iovisor/iovisor-ovn/iomodules/dhcp"
	"github.com/iovisor/iovisor-ovn/iomodules/l2switch"
	"github.com/iovisor/iovisor-ovn/iomodules/nat"
	"github.com/iovisor/iovisor-ovn/iomodules/null"
	"github.com/iovisor/iovisor-ovn/iomodules/onetoonenat"
	"github.com/iovisor/iovisor-ovn/iomodules/router"

	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("service-topology")

var dataplane *hover.Client

// List of deployed modules (Indexed by module name)
var modules map[string]iomodules.IoModule

// List of configuration for modules (Indexed by module name)
var modulesConfig map[string]interface{}

func deployModules(modulesRequested []interface{}) error {
	for _, i := range modulesRequested {
		iMap := to.Map(i)

		name_, nameok := iMap["name"]
		if !nameok {
			return errors.New("name is missing in module")
		}
		name := name_.(string)

		mtype_, mtypeok := iMap["type"]
		if !mtypeok {
			errString := fmt.Sprintf("type is missing in module '%s'", name)
			return errors.New(errString)
		}
		mtype := mtype_.(string)
		config_, configok := iMap["config"]
		if !configok {
			log.Warningf("config is missing in module '%s'", name)
		}
		config := to.Map(config_)

		var m iomodules.IoModule

		switch mtype {
		case "dhcp":
			m = dhcp.Create(dataplane)
		case "router":
			m = router.Create(dataplane)
		case "switch":
			m = l2switch.Create(dataplane)
		case "nat":
			m = nat.Create(dataplane)
		case "onetoonenat":
			m = onetoonenat.Create(dataplane)
		case "null_node": // "null" can not be used because it causes an issue
			m = null.Create(dataplane)
		default:
			errString := fmt.Sprintf("Invalid module type '%s' for module '%s'",
				mtype, name)
			return errors.New(errString)
		}

		if err := m.Deploy(); err != nil {
			errString := fmt.Sprintf("Error deploying module '%s': %s", name, err)
			return errors.New(errString)
		}

		modules[name] = m
		if configok {
			modulesConfig[name] = config
		}
	}

	return nil
}

func linkModules(linksRequested []interface{}) error {
	for _, i := range linksRequested {
		iMap := to.Map(i)
		from_, fromok := iMap["from"]
		if !fromok {
			return errors.New("from is not present on link request")
		}
		from := from_.(string)

		to_, took := iMap["to"]
		if !took {
			return errors.New("to is not present on link request")
		}
		to := to_.(string)

		log.Noticef("Linking modules: '%s' -> '%s'", to, from)

		_, ok1 := modules[to]
		if !ok1 {
			errString := fmt.Sprintf("Module '%s' does not exist", to)
			return errors.New(errString)
		}

		_, ok2 := modules[from]
		if !ok2 {
			errString := fmt.Sprintf("Module '%s' does not exist", from)
			return errors.New(errString)
		}

		// this call is a little trickly, it receives
		// (dataplame, iomodule1, name1, iomodule2, name2), by definition if two
		// modules are connected together the name of those interface on each
		// module is the name of the peer module.
		err := iomodules.AttachIoModules(dataplane, modules[from], to, modules[to], from)
		if err != nil {
			log.Error(err)
			errString := fmt.Sprintf("Error linking Modules '%s' -> '%s'", from, to)
			return errors.New(errString)
		}
	}

	return nil
}

func linkModulesToInterfaces(links []interface{}) error {
	for _, i := range links {
		iMap := to.Map(i)

		module_, moduleok := iMap["module"]
		if !moduleok {
			return errors.New("'module' is missing in link to interface")
		}
		module := module_.(string)

		iface_, ifaceok := iMap["iface"]
		if !ifaceok {
			return errors.New("'iface' is missing in link to interface")
		}
		iface := iface_.(string)

		m, mok := modules[module]
		if !mok {
			errString := fmt.Sprintf("Module '%s' does not exist", module)
			return errors.New(errString)
		}

		err := m.AttachExternalInterface(iface)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
	}

	return nil
}

func undeployModules() {
	for name, m := range modules {
		err := m.Destroy()
		if err != nil {
			log.Errorf("Error destroying module '%s': %s", name, err)
			// In this case it is a bad idea to stop, it is better to try to
			// remove all the modules
		}
	}
}

func DeployTopology(path string) error {

	dataplane = hover.NewClient()

	// Connect to hover and initialize HoverDataplane
	if err := dataplane.Init(config.Hover); err != nil {
		log.Errorf("unable to conect to Hover in '%s': %s", config.Hover, err)
		return err
	}

	modules = make(map[string]iomodules.IoModule)
	modulesConfig = make(map[string]interface{})

	conf, err := yaml.Open(path)
	if err != nil {
		log.Errorf("Failed to open configuraton fail '%s'", path)
		return err
	}

	// deploy iomodules
	log.Noticef("Deploying Modules")
	modulesRequested := conf.Get("modules")
	if modulesRequested == nil {
		return errors.New("No modules present on tolology file")
	}
	if err := deployModules(to.List(modulesRequested)); err != nil {
		log.Errorf("Error deploying modules: %s", err)
		// TODO: undeploy modules
		return err
	}

	// link modules together
	log.Noticef("Linking Modules")
	linksRequested := conf.Get("links")
	if linksRequested != nil {
		if err := linkModules(to.List(linksRequested)); err != nil {
			log.Errorf("%s", err)
			// TODO: remove links and undeploy modules
			return err
		}
	}

	// link modules to external interfaces
	log.Noticef("Connecting external interfaces to Modules")
	externalInterfacesRequested := conf.Get("external_interfaces")
	if externalInterfacesRequested != nil {
		if err := linkModulesToInterfaces(to.List(externalInterfacesRequested)); err != nil {
			log.Errorf("%s", err)
			// TODO: remove links and undeploy modules
			return err
		}
	}

	// configure modules
	log.Noticef("Configuring Modules")
	for name, m := range modules {
		if conf, ok := modulesConfig[name]; ok {
			err := m.Configure(conf)
			if err != nil {
				log.Errorf("Error configuring Module '%s': %s", name, err)
				return err
			}
		} else {
			log.Warningf("sikipping module '%s' without configuration", name)
		}
	}

	return nil
}

func UndeployTopology() {
	log.Errorf("Not Implemented")
}
