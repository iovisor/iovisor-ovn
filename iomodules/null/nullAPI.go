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
package null

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/netgroup-polito/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-null")

type NullModule struct {
	ModuleId   string

	linkIdHover string
	ifaceName   string

	deployed    bool
	hc          *hover.Client // used to send commands to hover
}

func Create(hc *hover.Client) *NullModule {

	if hc == nil {
		log.Errorf("Dataplane is not valid")
		return nil
	}

	x := new(NullModule)
	x.hc = hc
	x.deployed = false
	return x
}

func (m *NullModule) GetModuleId() string {
	return m.ModuleId
}

func (m *NullModule) Deploy() (err error) {

	if m.deployed {
		return nil
	}

	nullError, nullHover := m.hc.ModulePOST("bpf", "null", null_code)
	if nullError != nil {
		log.Errorf("Error in POST null IOModule: %s\n", nullError)
		return nullError
	}

	log.Noticef("POST NULL IOModule %s\n", nullHover.Id)
	m.ModuleId = nullHover.Id
	m.deployed = true

	id, _ := strconv.Atoi(m.ModuleId[2:])
	m.hc.GetController().RegisterCallBack(uint16(id), m.ProcessPacket)

	return nil
}

func (m *NullModule) Destroy() (err error) {

	if !m.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := m.hc.ModuleDELETE(m.ModuleId)
	if moduleDeleteError != nil {
		log.Errorf("Error in destrying DHCP IOModule: %s\n", moduleDeleteError)
		return moduleDeleteError
	}

	m.ModuleId = ""
	m.deployed = false

	return nil
}

func (m *NullModule) AttachExternalInterface(ifaceName string) (err error) {

	if !m.deployed {
		errString := "Trying to attach port in undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	linkError, linkHover := m.hc.LinkPOST("i:"+ifaceName, m.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

	m.linkIdHover = linkHover.Id
	m.ifaceName = ifaceName

	return nil
}

func (m *NullModule) DetachExternalInterface(ifaceName string) (err error) {

	if !m.deployed {
		errString := "Trying to detach port in undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if m.ifaceName != ifaceName {
		errString := fmt.Sprintf("Iface '%s' is not present in module '%s'\n",
			ifaceName, m.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	linkDeleteError, _ := m.hc.LinkDELETE(m.linkIdHover)

	if linkDeleteError != nil {
		log.Warningf("Problem removing iface '%s' from module '%s'\n",
			ifaceName, m.ModuleId)
		return linkDeleteError
	}

	m.linkIdHover = ""
	m.ifaceName = ""
	return nil
}

func (m *NullModule) AttachToIoModule(ifaceId int, ifaceName string) (err error) {

	if !m.deployed {
		errString := "Trying to attach port in undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	m.ifaceName = ifaceName
	m.linkIdHover = ""

	return nil
}

func (m *NullModule) DetachFromIoModule(ifaceName string) (err error) {
	if !m.deployed {
		errString := "Trying to detach port in undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	m.linkIdHover = ""
	m.ifaceName = ""
	return nil
}

func (m *NullModule) Configure(conf interface{}) (err error) {
	_ = conf
	return nil
}

func (m *NullModule) ProcessPacket(p *hover.Packet) (err error) {
	_ = p

	log.Infof("Null: '%s': Packet arrived from dataplane", m.ModuleId)
	return nil
}

