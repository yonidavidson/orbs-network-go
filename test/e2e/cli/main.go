package main

import (
	"fmt"
	"github.com/orbs-network/orbs-network-go/test/builders"
	"github.com/orbs-network/orbs-network-go/test/e2e"
	"time"
)

func main() {
	tx := builders.TransferTransaction().WithAmount(17).Builder()

	fmt.Println("sending transaction", tx)
	fmt.Println(e2e.SendTransaction(tx))

	time.Sleep(5 * time.Second)
	fmt.Println(e2e.GetBalance())

}
