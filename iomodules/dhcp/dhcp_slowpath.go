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
	"errors"
	"net"
	"strconv"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"

	"github.com/iovisor/iovisor-ovn/hover"
)

func (m *DhcpModule) ProcessPacket(p *hover.PacketIn) (err error) {
	m.c <- p
	return nil
}

// the dhcp library needs an object that implements the dhcp.ServeConn to
// be able to receive and send packets.  In this case this interface is
// directly implemented in the dhcp module.

// This function is called from the dhcp library, it waits on a channel
// until the ProcessPacket function injects new arrived packets there.
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

// WriteTo in this case means send the packet to the dataplane, it requires
// assemble the whole packet layers
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
