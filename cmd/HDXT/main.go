package main

import "ConjunctiveSSE/HDXT"

func main() {
	db, err := HDXT.MySQLSetup("Crime_USENIX_REV")
	if err != nil {
		panic(err)
	}

	db.Create(&HDXT.CipherText{
		Address: "0x123",
		Value:   "0x123",
	})
	db.Create(&HDXT.CipherText{
		Address: "0x124",
		Value:   "0x124",
	})
}
