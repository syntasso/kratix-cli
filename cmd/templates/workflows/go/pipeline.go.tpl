package main

import (
	"fmt"

	kratix "github.com/syntasso/kratix-go"
)

func main() {
	sdk := kratix.New()
	if sdk.WorkflowType() == "promise" {
		fmt.Printf("Hello from %s", sdk.PromiseName())
	} else {
		resource, _ := sdk.ReadResourceInput()
		fmt.Printf("Hello from %s %s", resource.GetName(), resource.GetNamespace())
	}
}
