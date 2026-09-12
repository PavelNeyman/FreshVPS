package main

import "testing"

func TestSetState(t *testing.T) {
	setState(1, "wait_vpn_add_name", "")
	if chatState[1] != "wait_vpn_add_name" {
		t.Fatal(chatState[1])
	}
	setState(1, "", "")
	if chatState[1] != "" {
		t.Fatal("cleared")
	}
	setState(2, "wait_node_newname", "id-1")
	if chatExtra[2] != "id-1" {
		t.Fatal(chatExtra[2])
	}
}
