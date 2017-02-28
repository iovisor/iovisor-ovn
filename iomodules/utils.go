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

package iomodules

import (
	"bytes"
	"fmt"
	"net"
)

func MacToHexadecimalString(mac net.HardwareAddr) string {
	s := mac.String()

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

func MacToHexadecimalStringBigEndian(mac net.HardwareAddr) string {
	s := mac.String()

	var buffer bytes.Buffer
	buffer.WriteString("0x")
	buffer.WriteString(s[15:17])
	buffer.WriteString(s[12:14])
	buffer.WriteString(s[9:11])
	buffer.WriteString(s[6:8])
	buffer.WriteString(s[3:5])
	buffer.WriteString(s[0:2])

	return buffer.String()
}

func IpToHex(ip net.IP) string {
	if ip.To4() != nil {
		ba := []byte(ip.To4())
		ipv4HexStr := fmt.Sprintf("0x%02x%02x%02x%02x",
			ba[0], ba[1], ba[2], ba[3])
		return ipv4HexStr
	}

	return ""
}

func IpToHexBigEndian(ip net.IP) string {
	if ip.To4() != nil {
		ba := []byte(ip.To4())
		ipv4HexStr := fmt.Sprintf("0x%02x%02x%02x%02x",
			ba[3], ba[2], ba[1], ba[0])
		return ipv4HexStr
	}

	return ""
}

func ParseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}
