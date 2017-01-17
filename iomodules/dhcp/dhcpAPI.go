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
package dhcp

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/mvbpolito/gosexy/to"

	//"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	l "github.com/op/go-logging"


)

var log = l.MustGetLogger("iomodules-dhcp")

type DhcpModule struct {
	ModuleId   string

	linkIdHover string
	ifaceName   string

	mac         net.HardwareAddr
	ip         	net.IP
	dns         net.IP
	router      net.IP
	leaseTime   uint32

	deployed  bool
	dataplane *hoverctl.Dataplane // used to send commands to hover
}

func Create(dp *hoverctl.Dataplane) *DhcpModule {

	if dp == nil {
		log.Errorf("Daplane is not valid\n")
		return nil
	}

	x := new(DhcpModule)
	x.dataplane = dp
	x.deployed = false
	return x
}

func (m *DhcpModule) GetModuleId() string {
	return m.ModuleId
}

func (m *DhcpModule) Deploy() (err error) {

	if m.deployed {
		return nil
	}

	dhcpError, dhcpHover := hoverctl.ModulePOST(m.dataplane, "bpf",
		"DHCP", DhcpServer)
	if dhcpError != nil {
		log.Errorf("Error in POST Switch IOModule: %s\n", dhcpError)
		return dhcpError
	}

	log.Noticef("POST DHCP IOModule %s\n", dhcpHover.Id)
	m.ModuleId = dhcpHover.Id
	m.deployed = true

	return nil
}

func (m *DhcpModule) Destroy() (err error) {

	if !m.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := hoverctl.ModuleDELETE(m.dataplane, m.ModuleId)
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

	linkError, linkHover := hoverctl.LinkPOST(m.dataplane, "i:"+ifaceName, m.ModuleId)
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

	linkDeleteError, _ := hoverctl.LinkDELETE(m.dataplane, m.linkIdHover)

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
func (m *DhcpModule) ConfigureParameters(pool net.IPNet, dns net.IP, router net.IP,
	leaseTime uint32, serverMac net.HardwareAddr, serverIp net.IP) (err error) {
	if !m.deployed {
		errString := "Trying to configure undeployed module"
		log.Errorf(errString)
		return errors.New(errString)
	}

	// set module configuration
	subnetMask := "0x" + pool.Mask.String()
	dnsHex := ipToHex(dns)
	leaseTimeStr := strconv.FormatUint(uint64(leaseTime), 10)
	routerHex := ipToHex(router)
	serverIpHex := ipToHex(serverIp)
	serverMacHex := macToHexadecimalString(serverMac.String())

	var toSend string
	toSend = "{" + subnetMask + " " + dnsHex + " " + leaseTimeStr + " " +
		routerHex + " " + serverIpHex + " " + serverMacHex + "}"

	hoverctl.TableEntryPUT(m.dataplane, m.ModuleId, "config",
		"0", toSend)

	// Unfortunately until now only 10 IP addresses are allowed by server
	hosts, _ := getHosts(pool)
	for i := 0; i < 10; i++ {

		// set address pool
		ipHex := ipToHex(hosts[i])
		index := strconv.Itoa(i)
		toSend := "{" + ipHex + " 0 0 0}" // Status, time, mac
		hoverctl.TableEntryPUT(m.dataplane, m.ModuleId, "pool",
			index, toSend)

		hoverctl.TableEntryPUT(m.dataplane, m.ModuleId, "ip_to_index",
			ipHex, index)
	}

	m.mac = serverMac;
	m.ip = serverIp;
	m.dns = dns;
	m.router = router;
	m.leaseTime = leaseTime;

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

	pool_, ok1 := confMap["pool"]
	dns_, ok2 := confMap["dns"]
	gw_, ok3 := confMap["gw"]
	lease_time_ , ok4 := confMap["lease_time"]
	server_ip_ , ok5 := confMap["server_ip"]
	server_mac_ , ok6 := confMap["server_mac"]

	// TODO: some of these fields could be optional and have a default value
	if !ok1 {
		return errors.New("Missing pool")
	}

	if !ok2 {
		return errors.New("Missing dns")
	}

	if !ok3 {
		return errors.New("Missing gw")
	}

	if !ok4 {
		return errors.New("Missing lease_time")
	}

	if !ok5 {
		return errors.New("Missing server_ip")
	}

	if !ok6 {
		return errors.New("Missing server_mac")
	}

	_, pool, _ := net.ParseCIDR(pool_.(string))
	dns := net.ParseIP(dns_.(string))
	gw := net.ParseIP(gw_.(string))
	//temp , _ := strconv.ParseUint(lease_time_.(string), 10, 32)
	var lease_time uint32 = uint32(lease_time_.(int))
	mac_server, _ := net.ParseMAC(server_mac_.(string))
	ip_server := net.ParseIP(server_ip_.(string))

	return m.ConfigureParameters(*pool, dns, gw, lease_time, mac_server, ip_server)
}

// TODO: this function should be smarter
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

func ipToHex(ip net.IP) string {
	if ip.To4() != nil {
		ba := []byte(ip.To4())
		ipv4HexStr := fmt.Sprintf("0x%02x%02x%02x%02x", ba[0], ba[1], ba[2], ba[3])
		return ipv4HexStr
	}

	return ""
}

// taken from https://gist.github.com/kotakanbe/d3059af990252ba89a82
 func getHosts(ipnet net.IPNet) ([]net.IP, error) {
	var ips []net.IP
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		temp := make(net.IP, len(ip))
		copy(temp, ip)
		ips = append(ips, temp)
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
