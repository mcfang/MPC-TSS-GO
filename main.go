package main

import (
	"fmt"

	"github.com/bnb-chain/tss-lib/tss"
)

func main() {
	fmt.Println("Hello, Go!")

	parties := make([]*tss.Party, 3)

	for i := range parties {
		fmt.Println("party", i)
	}
}
