package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PavelNeyman/netductor/internal/addons"
)

func runAddons(args []string) {
	if len(args) == 0 || args[0] == "list" {
		b, _ := json.MarshalIndent(addons.ListAddons(), "", "  ")
		fmt.Println(string(b))
		return
	}
	switch args[0] {
	case "lampac":
		b, _ := json.MarshalIndent(addons.CollectLampac(), "", "  ")
		fmt.Println(string(b))
	default:
		fmt.Fprintln(os.Stderr, "usage: netductor addons [list|lampac]")
		os.Exit(2)
	}
}
