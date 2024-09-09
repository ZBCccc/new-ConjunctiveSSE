package main

import (
	"ConjunctiveSSE/ODXT"
	"flag"
	"fmt"
	"strings"
)

var testDbName, testPhase, testGroup string
var delRate int

func main() {
	// Parse command line arguments
	flag.StringVar(&testDbName, "db", "Crime_USENIX_REV", "database name")
	flag.StringVar(&testPhase, "phase", "c", "test phase")
	flag.StringVar(&testGroup, "group", "server", "test group")
	flag.IntVar(&delRate, "del_rate", 0, "delete rate")
	flag.Parse()

	// Run tests
	err := TestODXT()
	if err != nil {
		fmt.Println("TestODXT error:", err)
	}
}

func TestODXT() error {
	fmt.Println("*********************************************")
	fmt.Println("Test_on ", testDbName, "del_rate", delRate)
	fmt.Println("Start test_group", testGroup, "phase", testPhase)
	fmt.Println("Start initial db")
	var odxt ODXT.ODXT
	err := odxt.DBSetup(testDbName, false)
	if err != nil {
		fmt.Println("DBSetup error", err)
		return err
	}
	if strings.Contains(testPhase, "c") {
		odxt.CiphertextGenPhase(testDbName)
	} else if strings.Contains(testPhase, "s") {
		odxt.DeletionPhaseWithSearch(delRate)
	}

	return nil
}
