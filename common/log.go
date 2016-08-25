package common

import (
	"os"

	l "github.com/op/go-logging"
)

var format = l.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)

//Logger initialization
func LogInit() {
	// For demo purposes, create two backend for os.Stderr.
	backend1 := l.NewLogBackend(os.Stderr, "", 0)
	backend2 := l.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := l.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := l.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(l.CRITICAL, "")

	// Set the backends to be used.
	l.SetBackend(backend1Leveled, backend2Formatter)
}
