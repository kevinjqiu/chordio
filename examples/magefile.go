// +build mage

package main

import (
	"fmt"
	"sync"

	"github.com/magefile/mage/sh" // mg contains helpful utility functions, like Deps
)

const chordioBin = "../dist/chordio_linux_amd64/chordio"

func Cluster5() error {
	server := sh.RunCmd(chordioBin, "server")
	binds := []string{
		"127.0.0.1:10000",
		"127.0.0.1:11000",
		"127.0.0.1:12000",
		"127.0.0.1:13000",
		"127.0.0.1:14000",
	}

	wg := sync.WaitGroup{}

	for _, bind := range binds {
		wg.Add(1)
		go func(b string) {
			err := server("--bind", b, "--tracing.enabled", "false", "--rank", "15")
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}(bind)
	}

	wg.Wait()
	return nil
}
