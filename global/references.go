package global

import (
	"time"

	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
)

var Hh *ovnmonitor.HandlerHandler

var SleepTime = 3500 * time.Millisecond
