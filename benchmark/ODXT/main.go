package main

import (
	"ConjunctiveSSE/ODXT"
	"flag"
	"fmt"
	"strings"
)

var test_db_name, test_phase, test_group string
var del_rate int

func main() {
	// Parse command line arguments
	flag.StringVar(&test_db_name, "db", "ODXT", "database name")
	flag.StringVar(&test_phase, "phase", "setup", "test phase")
	flag.StringVar(&test_group, "group", "server", "test group")
	flag.IntVar(&del_rate, "del_rate", 0, "delete rate")
	flag.Parse()

	// Run tests
	err := TestODXT()
	if err != nil {
		fmt.Println("TestODXT error:", err)
	}
}

func TestODXT() error {
	fmt.Println("*********************************************")
	fmt.Println("Test_on ", test_db_name, "del_rate", del_rate)
	fmt.Println("Start test_group", test_group, "phase", test_phase)
	fmt.Println("Start initial db")
	err := ODXT.DBSetup(test_db_name)
	if err != nil {
		fmt.Println("DBSetup error", err)
		return err
	}
	if strings.Contains(test_phase, "c") {
		ODXT.CiphertextGenPhase()
	} else if strings.Contains(test_phase, "s") {
		ODXT.DeletionPhaseWithSearch(del_rate)
	}

	return nil
}
