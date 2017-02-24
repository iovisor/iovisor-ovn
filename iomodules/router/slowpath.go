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
package router

import (
	"encoding/binary"
	"encoding/hex"
	"net"
	"time"

	"github.com/foize/go.fifo"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/iovisor/iovisor-ovn/hover"
)

const SLOWPATH_ARP_REPLY = 1
const SLOWPATH_ARP_LOOKUP_MISS = 2

const CLEANUP_EVERY_N_PACKETS = 5

const MAX_ENTRY_AGE = 15 * time.Second
const MAX_QUEUE_LEN = 10

type BufferQueue struct {
	queue       *fifo.Queue
	last_access time.Time
}

func (r *RouterModule) ProcessPacket(p *hover.PacketIn) (err error) {

	log.Infof("[router-%d]: Packet arrived from dataplane", p.Md.Module_id)
	log.Infof("[router-%d]: pkt_len(%d) port_id(%d) reason(%d)\n", p.Md.Module_id, p.Md.Packet_len, p.Md.Port_id, p.Md.Reason)
	log.Infof("[router-%d]: next_hop: %x out_port:%d out_port_ip:%x\n", p.Md.Module_id, p.Md.Metadata[0], p.Md.Metadata[1], p.Md.Metadata[2])

	log.Debugf("[router-%d]: -----PKT RECEIVED----\n%s\n", p.Md.Module_id, hex.Dump(p.Data[0:p.Md.Packet_len]))

	//How to decode packets?
	packet := gopacket.NewPacket(p.Data[0:p.Md.Packet_len], layers.LayerTypeEthernet, gopacket.Default)

	// if ethlayer := packet.Layer(layers.LayerTypeEthernet); ethlayer != nil {
	// 	log.Infof("ETHERNET \n")
	// 	eth, _ := ethlayer.(*layers.Ethernet)
	// 	log.Infof("srcmac:%s dstmac:%s eth_type:%s \n", eth.SrcMAC.String(), eth.DstMAC.String(), eth.EthernetType.String())
	// }
	//
	// if arplayer := packet.Layer(layers.LayerTypeARP); arplayer != nil {
	// 	log.Infof("ARP \n")
	// 	arp, _ := arplayer.(*layers.ARP)
	// 	log.Infof("op: %d (%d-REQ %d-REPLY) src_ha:%s dst_ha:%s src_ip:%s dst_ip:%s\n", arp.Operation, layers.ARPRequest, layers.ARPReply, hex.EncodeToString(arp.SourceHwAddress), hex.EncodeToString(arp.DstHwAddress), hex.EncodeToString(arp.SourceProtAddress), hex.EncodeToString(arp.DstProtAddress))
	// }
	//
	// if iplayer := packet.Layer(layers.LayerTypeIPv4); iplayer != nil {
	// 	log.Infof("IP \n")
	// 	ip, _ := iplayer.(*layers.IPv4)
	// 	log.Infof("srcIP:%s dstIP:%s\n", ip.SrcIP.String(), ip.DstIP.String())
	// }
	//
	// if icmplayer := packet.Layer(layers.LayerTypeICMPv4); icmplayer != nil {
	// 	log.Infof("ICMP \n")
	// 	icmp, _ := icmplayer.(*layers.ICMPv4)
	// 	log.Infof("icmp seq_num:%d\n", icmp.Seq)
	// }

	switch p.Md.Reason {
	case SLOWPATH_ARP_LOOKUP_MISS:
		log.Infof("[router-%d]: Reason -> ARP_LOOKUP_MISS\n", p.Md.Module_id)

		//enqueue packet_out, indexed for next_hop.
		pkt_to_queue := hover.PacketOut{}
		pkt_to_queue.Data = p.Data[0:p.Md.Packet_len]
		pkt_to_queue.Module_id = p.Md.Module_id
		pkt_to_queue.Sense = hover.EGRESS
		port_out := uint16(p.Md.Metadata[1])
		pkt_to_queue.Port_id = port_out

		// log.Infof("generate pkt_out module_id:%d Egress port_out:%d\n%s\n", pkt_to_queue.Module_id, pkt_to_queue.Port_id, pkt_to_queue.Data)

		//init queue if not initialized
		if _, ok := r.OutputBuffer[p.Md.Metadata[0]]; !ok {
			//miss
			bufQueue := BufferQueue{}

			newqueue := fifo.NewQueue()
			bufQueue.queue = newqueue
			bufQueue.last_access = time.Now()
			r.OutputBuffer[p.Md.Metadata[0]] = &bufQueue
		}

		//lookup for queue of packets, indexed for next_hop_ip
		if bufQueue, ok := r.OutputBuffer[p.Md.Metadata[0]]; ok {
			//hit
			bufQueue.queue.Add(&pkt_to_queue)
			bufQueue.last_access = time.Now()
		}

		// Generating ARP Request (after pkt queueing)
		if ethlayer := packet.Layer(layers.LayerTypeEthernet); ethlayer != nil {
			eth, _ := ethlayer.(*layers.Ethernet)
			dstmac, _ := net.ParseMAC("ff:ff:ff:ff:ff:ff")

			arppkt, arperr := buildArpPacket(eth.SrcMAC, dstmac, int2ip(p.Md.Metadata[2]), int2ip(p.Md.Metadata[0]), layers.ARPRequest)
			if arperr != nil {
				log.Error(arperr)
				return nil
			}

			p_out := hover.PacketOut{}
			p_out.Data = arppkt
			p_out.Module_id = p.Md.Module_id
			p_out.Sense = hover.EGRESS
			port_out := uint16(p.Md.Metadata[1])
			p_out.Port_id = port_out

			log.Infof("[router-%d]: sending arp request on port (%d) 'who has %x ? tell %x' \n", p.Md.Module_id, p_out.Port_id, p.Md.Metadata[0], p.Md.Metadata[2])
			ctrl := r.hc.GetController()
			ctrl.SendPacketOut(&p_out)
		}

	case SLOWPATH_ARP_REPLY:
		log.Infof("[router-%d]: Reason -> ARP_REPLY\n", p.Md.Module_id)

		if arplayer := packet.Layer(layers.LayerTypeARP); arplayer != nil {
			arp, _ := arplayer.(*layers.ARP)
			//reply?
			if arp.Operation != layers.ARPReply {
				log.Warningf("[router-%d]: No arp reply received!\n", p.Md.Module_id)
			}

			log.Infof("[router-%d]: ARP reply '%s is at %s'\n", p.Md.Module_id, net.HardwareAddr(arp.SourceHwAddress).String(), net.IP(arp.SourceProtAddress).To4().String())

			next_hop_ip_int := ip2int(net.IP(arp.SourceProtAddress))

			if BufQueue, ok := r.OutputBuffer[next_hop_ip_int]; ok {
				// log.Infof("Output Buffer Lookup HIT (%x)\n", next_hop_ip_int)
				if BufQueue != nil {
					if BufQueue.queue.Len() == 0 {
						// log.Infof("Delete empty queue\n")
						delete(r.OutputBuffer, next_hop_ip_int)
					}
				}
				if BufQueue != nil {
					if BufQueue.queue.Len() > 0 {
						// log.Infof("Output queue not empty. Send Packets out ...\n")
						//send out packets enqueued
						for BufQueue.queue.Len() > 0 {
							item := BufQueue.queue.Next()
							pkt := item.(*hover.PacketOut)
							packet := gopacket.NewPacket(pkt.Data, layers.LayerTypeEthernet, gopacket.Default)

							if ethlayer := packet.Layer(layers.LayerTypeEthernet); ethlayer != nil {
								eth, _ := ethlayer.(*layers.Ethernet)
								eth.DstMAC = arp.SourceHwAddress

								buf := gopacket.NewSerializeBuffer()
								opts := gopacket.SerializeOptions{}
								gopacket.SerializeLayers(buf, opts, eth, gopacket.Payload(pkt.Data[14:]))

								log.Infof("[router-%d]: send out packet from buffer. (current buffer size = %d)\n", p.Md.Module_id, BufQueue.queue.Len())
								pkt.Data = buf.Bytes()

								// log.Infof("Sending PacketOUT\n")
								ctrl := r.hc.GetController()
								ctrl.SendPacketOut(pkt)
							}
						}
					}
				}
			}
		}
	}
	/* CLEANUP */
	r.PktCounter++
	//Perform OutputBuffer Cleanup every CLEANUP_EVERY_N_PACKETS received by slowpath
	if r.PktCounter%CLEANUP_EVERY_N_PACKETS == 0 {
		log.Infof("[router-%d]: Perform cleanup.\n", p.Md.Module_id)
		for next_hop_ip, outbuffer := range r.OutputBuffer {
			age := time.Now().Sub(outbuffer.last_access)
			// Delete all queue with last_access older than MAX_ENTRY_AGE
			if age > MAX_ENTRY_AGE {
				// Free queue
				log.Infof("[router-%d]: queue [%x] (size=%d) age %s > %s (MAX_AGE)\n", p.Md.Module_id, next_hop_ip, outbuffer.queue.Len(), age, MAX_ENTRY_AGE)
				for outbuffer.queue.Len() > 0 {
					outbuffer.queue.Next()
				}
				// Delete map entry
				delete(r.OutputBuffer, next_hop_ip)
			} else {
				// Delete old entries, maintain max queue size to MAX_QUEUE_LEN
				if outbuffer.queue.Len() > MAX_QUEUE_LEN {
					log.Infof("[router-%d]: queue [%x] size %d > %d (MAX_SIZE)\n", p.Md.Module_id, next_hop_ip, outbuffer.queue.Len(), MAX_QUEUE_LEN)
					for outbuffer.queue.Len() > MAX_QUEUE_LEN {
						outbuffer.queue.Next()
					}
					log.Infof("[router-%d]: queue [%x] size %d == %d (MAX_SIZE) CLEANUP ENDED...\n", p.Md.Module_id, next_hop_ip, outbuffer.queue.Len(), MAX_QUEUE_LEN)
				}
			}
			// Delete Map entries with empty queue
			if outbuffer.queue.Len() == 0 {
				log.Infof("[router-%d]: queue [%x] size %d DELETING...\n", p.Md.Module_id, next_hop_ip, outbuffer.queue.Len())
				delete(r.OutputBuffer, next_hop_ip)
			}
		}
	}
	return nil
}

func buildArpPacket(srcmac net.HardwareAddr, dstmac net.HardwareAddr, srcip net.IP, dstip net.IP, operation uint16) ([]byte, error) {
	eth, arp, err := buildArpPacketLayers(srcmac, dstmac, srcip, dstip, operation)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{}
	gopacket.SerializeLayers(buf, opts, &eth, &arp)
	// log.Debugf("\nGenerated ARP Packet\n%s\n", hex.Dump(buf.Bytes()))

	return buf.Bytes(), nil
}

// buildArpPacket creates an template ARP packet with the given source and
// destination.
func buildArpPacketLayers(srcmac net.HardwareAddr, dstmac net.HardwareAddr, srcip net.IP, dstip net.IP, operation uint16) (layers.Ethernet, layers.ARP, error) {
	ether := layers.Ethernet{
		EthernetType: layers.EthernetTypeARP,

		SrcMAC: srcmac,
		DstMAC: dstmac,
	}
	arp := layers.ARP{
		AddrType: layers.LinkTypeEthernet,
		Protocol: layers.EthernetTypeIPv4,

		Operation: operation,

		HwAddressSize:   6,
		ProtAddressSize: 4,

		SourceHwAddress:   []byte(srcmac),
		SourceProtAddress: []byte(srcip.To4()),

		DstHwAddress:   []byte(dstmac),
		DstProtAddress: []byte(dstip.To4()),
	}
	return ether, arp, nil
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}
