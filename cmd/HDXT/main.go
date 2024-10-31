package main

import (
	"ConjunctiveSSE/pkg/HDXT"
)

func main() {
	db, err := HDXT.MySQLSetup("Crime_USENIX_REV")
	if err != nil {
		panic(err)
	}

	db.Create(&HDXT.MitraCipherText{
		Address: "0x123",
		Value:   "0x123",
	})
	db.Create(&HDXT.MitraCipherText{
		Address: "0x124",
		Value:   "0x124",
	})
}
