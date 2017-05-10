package main

import (
	"fmt"
	// "log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"apiServer/server"
	"apiServer/config"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU()) //

	// sessionCh := make(chan bool)

	config.Init()

	go server.Start(config.C_ExCfg.ServerAddr)
	fmt.Println("===server", config.C_ExCfg.ServerAddr, " start complated ===")

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGTERM)
	for {
		msg := <-ch
		switch msg {
		case syscall.SIGHUP:
		case syscall.SIGTERM:
			// sessionCh <- true
			fmt.Println("===server closed===")
			os.Exit(0)
		}
	}

	fmt.Println("============server closed ============")
}
