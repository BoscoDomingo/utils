package main

import (
	"fmt"
	"os"

	"github.com/BoscoDomingo/utils/go/tools/zcp/internal/zcp"
)

func main() {
	if err := zcp.Run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "zcp: %v\n", err)
		os.Exit(1)
	}
}
