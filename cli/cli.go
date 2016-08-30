package cli

import (
	"bufio"
	"fmt"
	"os"

	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/testenv"
)

func Cli(dataplane *hoverctl.Dataplane) {
	for i := 0; i < 10000; i++ {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		text, _ := reader.ReadString('\n')
		fmt.Println(text)

		switch text {
		case "test\n":
			testenv.TestSwitch2ifc(dataplane, "i:veth1_", "i:veth2_")
		default:
			fmt.Println("invalid command")
		}
	}
}
