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

var null_code = `
#include <bcc/proto.h>
#include <bcc/helpers.h>

#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/if_packet.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/in.h>
#include <uapi/linux/filter.h>
#include <uapi/linux/pkt_cls.h>

static int handle_rx(void *skb, struct metadata *md) {

  bpf_trace_printk("[null-%d]: in_fc: %d\n", md->module_id, md->in_ifc);

  //return RX_DROP;

  //pkt_redirect(skb, md, md->in_ifc);
  //return RX_REDIRECT;

  //pkt_controller(skb, md, 700);
  return RX_CONTROLLER;
}
`
