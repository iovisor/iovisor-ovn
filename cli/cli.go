package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/testenv"
)

func Cli(dataplane *hoverctl.Dataplane) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("cli@iov-ovn$")
		text, _ := reader.ReadString('\n')

		switch text {
		case "test\n":
			testenv.TestSwitch2ifc(dataplane, "i:veth1_", "i:veth2_")
		case "\n":
		default:
			fmt.Println("invalid command")
		}
	}
}
