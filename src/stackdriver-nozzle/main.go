package main

import (
)
import "github.com/evandbrown/gcp-tools-boshrelease/src/stackdriver-nozzle/sink"

func main() {
	client := sink.NewStackdriverClient()
	client.Post("Hello world again")
}
