package ODXT

import (
	"testing"
)

func TestODXT(t *testing.T) {
	var client Client
	var server Server

	// ODXTClient.Setup
	err := client.Setup()
	if err != nil {
		t.Error(err)
		return
	}

	// ODXTServer.Setup
	server.Setup()

	// ODXTClient.Update
}
