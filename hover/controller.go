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
package hover

import (
	"encoding/gob"
	"fmt"
	"net"
	"io"
)

type SlowPathCallBack func(*Packet) (error)

type Packet struct {
	Module_id  uint16
	Port_id    uint16
	Packet_len uint16
	Reason     uint16
	Data       []byte
}

const (
	INGRESS = 0
	EGRESS = 1
)

type PacketOut struct {
	Module_id  uint16
	Port_id    uint16
	Sense      uint16 /* ingress = 0, egress = 1 */
	Data       []byte
}

func (p *Packet) ToString() string {
	return fmt.Sprintf("Module_id: %d\nPort_id: %d\nPacket_len: %d\nReason: %d\n",
		p.Module_id, p.Port_id, p.Packet_len, p.Reason)
}

type Controller struct {
	callbacks  map[uint16]SlowPathCallBack
	listenaddr string
	conn       net.Conn
}

func (c *Controller) Init(listenaddr string) (err error) {
	c.callbacks = make(map[uint16]SlowPathCallBack)
	c.listenaddr = listenaddr
	return nil
}

func (c *Controller) RegisterCallBack(id uint16, cb SlowPathCallBack) (err error) {
	c.callbacks[id] = cb
	return nil
}

func (c *Controller) Run() (err error) {
	ln, err1 := net.Listen("tcp", c.listenaddr)
	if err1 != nil {
		return err1
	}

	c.conn, err = ln.Accept()
	if err != nil {
		return
	}

	dec := gob.NewDecoder(c.conn)

	log.Infof("Controller: New client connected")

	for {
		p := &Packet{}
		err1 := dec.Decode(p)
		if err1 == io.EOF {
			log.Info("Controller: Client Disconnected")
			return nil
		} else if err1 != nil {
			continue
		}
		//log.Info("Packet arrived: ", p.ToString())
		cb, ok := c.callbacks[p.Module_id]
		if !ok {
			log.Infof("Controller: Callback not found")
			continue
		}
		cb(p)
	}
}

func (c *Controller) SendPacketOut(p *PacketOut) (err error) {
	encoder := gob.NewEncoder(c.conn)
	encoder.Encode(p)
	return
}
