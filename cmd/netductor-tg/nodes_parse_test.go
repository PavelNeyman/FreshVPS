package main

import "testing"

func TestParseNodesListLine(t *testing.T) {
	// exercise format with empty exec by ensuring empty list doesn't panic
	_ = formatNodesListHTML()
	_ = nodesRenameKeyboard()
}
