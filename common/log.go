package common

import (
	"os"

	"github.com/netgroup-polito/iovisor-ovn/config"
	l "github.com/op/go-logging"
)

// var format = l.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)
var format = l.MustStringFormatter(`%{color}%{time:15:04:05} %{level:.4s} >%{color:reset} %{message}`)

//Logger initialization
func LogInit() {

	backend := l.NewLogBackend(os.Stderr, "", 0)

	if config.Debug == true {
		format = l.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)
	}

	backendFormatter := l.NewBackendFormatter(backend, format)

	backendLeveled := l.SetBackend(backendFormatter)

	backendLeveled.SetLevel(l.NOTICE, "")

	if config.Info == true {
		backendLeveled.SetLevel(l.INFO, "")
	}

	if config.Debug == true {
		backendLeveled.SetLevel(l.DEBUG, "")
	}

	l.SetBackend(backendLeveled)
}
