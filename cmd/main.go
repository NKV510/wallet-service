package main

import (
	"fmt"

	"github.com/NKV510/wallet-service/internal"
)

func main() {
	cfg, _ := internal.Load()
	fmt.Println(cfg)
}
