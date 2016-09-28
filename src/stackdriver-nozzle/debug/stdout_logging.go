package debug

import "fmt"

type StdOut struct {
}

func (so *StdOut) Connect() bool {
	return true
}

func (so *StdOut) ShipEvents(event map[string]interface{}, msg string) {
	if msg != "" {
		fmt.Printf("%s: %+v\n\n", msg, event)
	}
}

