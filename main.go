package main

import (
	"fmt"
	"math/big"
)

func main() {
	a := big.NewInt(4)
	b := big.NewInt(11)

	g := big.NewInt(65537)
	//p, _ := new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)

	fmt.Println(a.ModInverse(a, b))
	fmt.Println(a.ModInverse(g, b))

}
