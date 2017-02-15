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
package iomodules

import (
	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules")

type IoModule interface {
	GetModuleId() string
	Deploy() (err error)
	Destroy() (err error)
	AttachExternalInterface(name string) (err error)
	DetachExternalInterface(name string) (err error)
	AttachToIoModule(IfaceId int, name string) (err error)
	DetachFromIoModule(name string) (err error)
	Configure(conf interface{}) (err error)
	ProcessPacket(p *hover.Packet) (err error)
}

// this function attaches to modules together.  It performs the reques to hover
// and then it calls the AttachToIoModule of each module to register the interface
func AttachIoModules(c *hover.Client,
	m1 IoModule, ifaceName1 string, m2 IoModule, ifaceName2 string) (err error) {

	link_err, link := c.LinkPOST(m1.GetModuleId(), m2.GetModuleId())
	if link_err != nil {
		log.Errorf("%s", link_err)
		return
	}

	// hover does not guarantee that the order of the link is conserved, then
	// it is necessary to check it explicitly to realize the interface id
	// inside each module
	m1id := -1
	m2id := -1

	if link.From == m1.GetModuleId() {
		m1id = link.FromId
		m2id = link.ToId
	} else {
		m1id = link.ToId
		m2id = link.FromId
	}

	if err := m1.AttachToIoModule(m1id, ifaceName1); err != nil {
		log.Errorf("%s", err)
		return err
	}

	if err := m2.AttachToIoModule(m2id, ifaceName2); err != nil {
		log.Errorf("%s", err)
		return err
	}

	return nil
}
