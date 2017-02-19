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
package dhcp_user

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	// We use this packet that is a fork of "github.com/krolaw/dhcp4"
	// to avoid any problem due to an API change
	dhcp "github.com/mvbpolito/dhcp4"

	"github.com/mvbpolito/gosexy/to"

	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-dhcp-user")

type DhcpModule struct {
	ModuleId   string

	linkIdHover string
	ifaceName   string

	mac         net.HardwareAddr
	ip          net.IP

	handler     dhcp.Handler
	c           chan *hover.Packet

	deployed  bool
	hc *hover.Client // used to send commands to hover
}

func Create(hc *hover.Client) *DhcpModule {

	if hc == nil {
		log.Errorf("HoverClient is not valid")
		return nil
	}

	x := new(DhcpModule)
	x.hc = hc
	x.deployed = false
	x.c = make(chan *hover.Packet, 10)
	return x
}

func (m *DhcpModule) GetModuleId() string {
	return m.ModuleId
}

func (m *DhcpModule) Deploy() (err error) {

	if m.deployed {
		return nil
	}

	dhcpError, dhcpHover := m.hc.ModulePOST("bpf", "DHCP", DhcpServer)
	if dhcpError != nil {
		log.Errorf("Error in POST dhcp IOModule: %s\n", dhcpError)
		return dhcpError
	}

	log.Noticef("POST DHCP IOModule %s\n", dhcpHover.Id)
	m.ModuleId = dhcpHover.Id
	m.deployed = true

	id, _ := strconv.Atoi(m.ModuleId[2:])
	m.hc.GetController().RegisterCallBack(uint16(id), m.ProcessPacket)

	return nil
}

func (m *DhcpModule) Destroy() (err error) {

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

func (m *DhcpModule) AttachExternalInterface(ifaceName string) (err error) {

	if !m.deployed {
		errString := "Trying to attach port in undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if m.ifaceName != "" {
		errString := fmt.Sprintf("Module '%s' is already connected to interface '%s'\n",
			m.ModuleId, ifaceName)
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

func (m *DhcpModule) DetachExternalInterface(ifaceName string) (err error) {

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

func (m *DhcpModule) AttachToIoModule(ifaceId int, ifaceName string) (err error) {

	if !m.deployed {
		errString := "Trying to attach port in undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if m.ifaceName != "" {
		errString := fmt.Sprintf("Module '%s' is already connected to interface '%s'\n",
			m.ModuleId, ifaceName)
		log.Errorf(errString)
		return errors.New(errString)
	}

	m.ifaceName = ifaceName
	m.linkIdHover = ""

	return nil
}

func (m *DhcpModule) DetachFromIoModule(ifaceName string) (err error) {
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

	m.linkIdHover = ""
	m.ifaceName = ""
	return nil
}

// TODO: this function should be split on smaller pieces.
func (m *DhcpModule) ConfigureParameters(netmask net.IPMask,
										addr_low net.IP,
										addr_high net.IP,
										dns net.IP,
										router net.IP,
										leaseTime uint32,
										serverMAC net.HardwareAddr,
										serverIP net.IP) (err error) {
	if !m.deployed {
		errString := "Trying to configure undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	m.handler = &DHCPHandler{
		ip:            serverIP[12:16],
		leaseDuration: time.Duration(leaseTime)*time.Second,
		start:         addr_low,
		leaseRange:    dhcp.IPRange(addr_low, addr_high),
		leases:        make(map[int]lease, 10), // TODO: what is this "10" for?
		options: dhcp.Options{
			dhcp.OptionSubnetMask:       []byte(netmask),
			dhcp.OptionRouter:           []byte(router[12:16]),
			dhcp.OptionDomainNameServer: []byte(dns[12:16]),
		},
	}

	m.mac = serverMAC
	m.ip = serverIP

	// mac and ip addresses are used in the dataplane to decide if a packet
	// has to be sent to the controller or not.
	serverIpHex := ipToHex(serverIP)
	serverMacHex := macToHexadecimalString(serverMAC.String())

	var toSend string
	toSend = "{" + serverIpHex + " " + serverMacHex + "}"

	m.hc.TableEntryPUT(m.ModuleId, "config", "0", toSend)

	go dhcp.Serve(m, m.handler)

	return nil
}

func (m *DhcpModule) Configure(conf interface{}) (err error) {
	// conf is a map that contains:
	//		pool: CIDR notation of the address pool
	// 		dns: ip address of the DNS given to clients
	//		gw: ip of the default routers given to clients
	//		lease_time: default lease_time
	//		server_ip: ip address of the dhcp server
	//		server_mac: mac address of the dhcp server

	log.Infof("Configuring DHCP server")
	confMap := to.Map(conf)

	netmask_, netmask_ok := confMap["netmask"]
	addr_low_, addr_low_ok := confMap["addr_low"]
	addr_high_, addr_high_ok := confMap["addr_high"]
	dns_, dns_ok := confMap["dns"]
	router_, router_ok := confMap["router"]
	lease_time_, lease_time_ok := confMap["lease_time"]
	server_ip_, server_ip_ok := confMap["server_ip"]
	server_mac_, server_mac_ok := confMap["server_mac"]

	// TODO: some of these fields could be optional and have a default value
	if !netmask_ok {
		return errors.New("Missing netmask")
	}

	if !addr_low_ok {
		return errors.New("Missing addr_low")
	}

	if !addr_high_ok {
		return errors.New("Missing addr_high")
	}

	if !dns_ok {
		return errors.New("Missing dns")
	}

	if !router_ok {
		return errors.New("Missing router")
	}

	if !lease_time_ok {
		return errors.New("Missing lease_time")
	}

	if !server_ip_ok {
		return errors.New("Missing server_ip")
	}

	if !server_mac_ok {
		return errors.New("Missing server_mac")
	}

	netmask := ParseIPv4Mask(netmask_.(string))
	addr_low := net.ParseIP(addr_low_.(string))
	addr_high := net.ParseIP(addr_high_.(string))
	dns := net.ParseIP(dns_.(string))
	router := net.ParseIP(router_.(string))
	var lease_time uint32 = uint32(lease_time_.(int))
	mac_server, _ := net.ParseMAC(server_mac_.(string))
	ip_server := net.ParseIP(server_ip_.(string))

	return m.ConfigureParameters(netmask, addr_low, addr_high, dns,
		router, lease_time, mac_server, ip_server)
}

func (m *DhcpModule) ProcessPacket(p *hover.Packet) (err error) {
	m.c <- p
	return nil
}

func (m *DhcpModule) ReadFrom(b []byte) (n int, addr net.Addr, err error) {

	for p := range m.c {
		packet := gopacket.NewPacket(p.Data, layers.LayerTypeEthernet, gopacket.Lazy)

		ethLayer := packet.Layer(layers.LayerTypeEthernet)
		if ethLayer == nil {
			log.Errorf("Error parsing packet: Ethernet")
			err = errors.New("Error parsing packet: Ethernet")
			continue
		}

		ipLayer := packet.Layer(layers.LayerTypeIPv4)
		if ipLayer == nil {
			log.Errorf("Error parsing packet: ipv4")
			err = errors.New("Error parsing packet: ipv4")
			continue
		}

		udpLayer := packet.Layer(layers.LayerTypeUDP)
		if udpLayer == nil {
			log.Errorf("Error parsing packet: udp")
			err = errors.New("Error parsing packet: udp")
			continue
		}

		eth, _ := ethLayer.(*layers.Ethernet)
		ip, _ := ipLayer.(*layers.IPv4)
		udp, _ := udpLayer.(*layers.UDP)

		_ = eth

		udpAddr := &net.UDPAddr{ip.SrcIP, int(udp.SrcPort), ""}
		addr = udpAddr

		copy(b, udp.LayerPayload())
		n = len(udp.LayerPayload())
		return
	}

	return
}

func (m *DhcpModule) WriteTo(b []byte, addr net.Addr) (n int, err error) {

	// TODO: For now all the packets are sent to the broadcast mac address,
	// this is ok for some dhclient implementations but it is not full c
	// complaint with the protocol specification
	eth := &layers.Ethernet {
		SrcMAC: m.mac,
		DstMAC: net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, // FIXME
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

func ParseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}

func ipToHex(ip net.IP) string {
	if ip.To4() != nil {
		ba := []byte(ip.To4())
		ipv4HexStr := fmt.Sprintf("0x%02x%02x%02x%02x", ba[0], ba[1], ba[2], ba[3])
		return ipv4HexStr
	}

	return ""
}

func macToHexadecimalString(s string) string {
	var buffer bytes.Buffer

	buffer.WriteString("0x")
	buffer.WriteString(s[0:2])
	buffer.WriteString(s[3:5])
	buffer.WriteString(s[6:8])
	buffer.WriteString(s[9:11])
	buffer.WriteString(s[12:14])
	buffer.WriteString(s[15:17])

	return buffer.String()
}
