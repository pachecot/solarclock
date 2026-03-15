package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

func usage() {
	exe := exePath()
	fmt.Printf(`solarclock is a web service for reporting solar information. 

Usage:

	%s <command> [arguments]

Commands:

	install     install service
	remove      remove service
	start       start service  	   	
	run         run server locally
	stop        stop service

Arguments:

	-port n
	--port n
	-p n
	            set the port of the http server 

`, exe)
}

type command int

const (
	install = iota
	remove
	start
	stop
	status
	run
	unknown
)

func parseCmd() command {
	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "install":
		return install
	case "remove":
		return remove
	case "start":
		return start
	case "status":
		return status
	case "run":
		return run
	case "stop":
		return stop
	default:
		return unknown
	}
}

const defaultPort = 8080

func getPort() uint32 {
	var port uint32 = defaultPort

	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		serviceParametersKey,
		registry.QUERY_VALUE)
	if err != nil {
		return port
	}
	value, vType, err := key.GetIntegerValue("port")
	if vType != registry.DWORD {
		return port
	}
	return uint32(value)
}

func setPort(port uint32) error {
	k, _, e := registry.CreateKey(
		registry.LOCAL_MACHINE,
		serviceParametersKey,
		registry.SET_VALUE|registry.QUERY_VALUE)
	if e != nil {
		return e
	}
	if port == defaultPort {
		k.DeleteValue("port")
		return nil
	}
	k.SetDWordValue("port", port)
	return nil
}

func parsePort() {
	if len(os.Args) < 4 {
		return
	}
	args := os.Args[2:]
	for i, arg := range args {
		opt := strings.ToLower(arg)
		switch opt {
		case "-p", "--port", "-port":
			if i+1 < len(args) {
				n, err := strconv.ParseUint(args[i+1], 10, 32)
				if err == nil {
					setPort(uint32(n))
					return
				}
			}
		}
	}
}

func main() {
	isService, err := svc.IsWindowsService()
	if err != nil {
		fmt.Printf("error %e\n", err)
		return
	}

	if isService {
		elog, err := eventlog.Open(serviceName)
		if err != nil {
			return
		}
		defer elog.Close()
		port := getPort()
		s := createClockService(port, elog)
		s.run()
		return
	}

	if len(os.Args) < 2 {
		usage()
		return
	}
	cmd := parseCmd()

	switch cmd {

	case install:
		parsePort()
		err := createService()
		if err != nil {
			fmt.Printf("error %e\n", err)
		}

	case remove:
		err := removeService()
		if err != nil {
			fmt.Printf("error %e\n", err)
		}

	case start:
		startService()

	case status:
		statusService()

	case run:
		elog := debug.New(serviceName)
		defer elog.Close()
		parsePort()
		port := getPort()
		s := createClockService(port, elog)
		s.runLocal()

	case stop:
		err := stopService()
		if err != nil {
			fmt.Printf("error %e\n", err)
		}
	}
}
