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
	"time"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-null")

type NullModule struct {
	ModuleId   string

	linkIdHover string
	ifaceName   string

	deployed    bool
	hc          *hover.Client // used to send commands to hover

	mac         net.HardwareAddr
	ip          net.IP
}

func Create(hc *hover.Client) *NullModule {

	if hc == nil {
		log.Errorf("Dataplane is not valid")
		return nil
	}

	x := new(NullModule)
	x.hc = hc
	x.deployed = false

	x.mac, _ = net.ParseMAC("6e:59:d3:54:75:2d")
	x.ip = net.ParseIP("8.8.8.7")
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

	//go m.sendPacketOut()

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

	packet := gopacket.NewPacket(p.Data, layers.LayerTypeEthernet, gopacket.Lazy)

	ethLayer := packet.Layer(layers.LayerTypeEthernet)
	if ethLayer == nil {
		log.Errorf("Error parsing packet: Ethernet")
		return errors.New("Error parsing packet: Ethernet")
	}

	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer == nil {
		log.Errorf("Error parsing packet: ipv4")
		return errors.New("Error parsing packet: ipv4")
	}

	udpLayer := packet.Layer(layers.LayerTypeUDP)
	if udpLayer == nil {
		log.Errorf("Error parsing packet: udp")
		return errors.New("Error parsing packet: udp")
	}

	eth, _ := ethLayer.(*layers.Ethernet)
	ip, _ := ipLayer.(*layers.IPv4)
	udp, _ := udpLayer.(*layers.UDP)

	_ = eth

	var addr net.UDPAddr

	addr.IP = ip.SrcIP
	addr.Port = int(udp.SrcPort)

	fmt.Println(len(udp.LayerPayload()))

	return nil
}

func (m *NullModule) sendPacketOut() {

	time.Sleep(time.Millisecond * 500)
	ticker := time.NewTicker(time.Millisecond * 1000)
	go func() {
		for _ = range ticker.C {
			addr := &net.UDPAddr{IP: net.IPv4bcast, Port: 67}
			m.WriteTo([]byte("Hola mama"), addr)
			//p := &hover.PacketOut{}
			//id, _ := strconv.Atoi(m.ModuleId[2:])
			//p.Module_id = uint16(id)
			//p.Port_id = 0
			//p.Sense = hover.INGRESS
			//p.Data = []byte("Here is a string....")
			//m.hc.GetController().SendPacketOut(p)

		}
	}()
}

func (m *NullModule) WriteTo(b []byte, addr net.Addr) (n int, err error) {

	// Create packet

	eth := &layers.Ethernet {
		SrcMAC: m.mac,
		DstMAC: net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // to FIX
		EthernetType: layers.EthernetTypeIPv4,
	}

	ipStr, portStr, err1 := net.SplitHostPort(addr.String())
	if err1 != nil {
		return
	}

	port, _ := strconv.Atoi(portStr)

	ip := &layers.IPv4{
		SrcIP: m.ip,
		DstIP: net.ParseIP(ipStr),
		Protocol: layers.IPProtocolUDP,
		Version: 4,
		TTL: 64,
	}

	udp := &layers.UDP{
		SrcPort: 68,
		DstPort: layers.UDPPort(port),
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths: true,
		ComputeChecksums: true,
	}

	udp.SetNetworkLayerForChecksum(ip)

	err = gopacket.SerializeLayers(buf, opts, eth, ip, udp, gopacket.Payload(b))
	if err != nil {
		log.Infof("Error in SerializeLayers: %s", err)
		return
	}

	p := &hover.PacketOut{}
	id, _ := strconv.Atoi(m.ModuleId[2:])
	p.Module_id = uint16(id)
	p.Port_id = 1
	p.Sense = hover.EGRESS
	p.Data = buf.Bytes()

	m.hc.GetController().SendPacketOut(p)

	return len(b), nil
}
