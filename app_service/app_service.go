// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/aknott2210/os_service/arguments"
	"github.com/kardianos/service"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

var logger service.Logger
var app string
var config string
var configFlag string
var serviceName string
var dependency string
var pid int

type program struct{}

func init() {
	flag.StringVar(&app, "app", "", "The path to the application.")
	flag.StringVar(&serviceName, "serviceName", "", "The name of the service.")
	flag.StringVar(&config, "config", "", "The path to the configuration.")
	flag.StringVar(&configFlag, "configFlag", "", "The configuration flag to provide to the application.")
	flag.StringVar(&dependency, "dependency", "", "Dependency to add to service start up.")
}

func init() {
	if !arguments.ServiceCall() {
		app = os.Args[1]
		config = os.Args[2]
		configFlag = os.Args[3]
		serviceName = os.Args[4]
		if len(os.Args) >= 6 {
			dependency = os.Args[5]
		}
	}
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logger.Info("Starting service...")
	go p.run()
	return nil
}

func (p *program) run() {
	cmd := exec.Command(app, "agent", configFlag, config)
	err := cmd.Start()
	pid = cmd.Process.Pid
	if err != nil {
		logger.Error(err)
	}
}

func (p *program) Stop(s service.Service) error {
	if win() {
		cmd := exec.Command("Taskkill", "/IM", strconv.Itoa(pid), "/F")
		err := cmd.Start()
		if err != nil {
			logger.Error(err)
		}
	}
	return nil
}

func win() bool {
	return runtime.GOOS == "windows"
}

func serviceConfig() *service.Config {
	if dependency == "" {
		return configNoDependency()
	} else {
		return configWithDependency()
	}
}

func configWithDependency() *service.Config {
	return &service.Config{
		Name:         serviceName,
		DisplayName:  serviceName,
		Description:  "This service starts up " + serviceName,
		Arguments:    []string{app, config, configFlag, serviceName},
		Dependencies: []string{dependency},
	}
}

func configNoDependency() *service.Config {
	return &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceName + " service",
		Arguments:   []string{app, config, configFlag, serviceName},
	}
}

func main() {
	svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	svcConfig := serviceConfig()

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	if len(*svcFlag) != 0 {
		err := service.Control(s, *svcFlag)
		if err != nil {
			log.Printf("Valid actions: %q\n", service.ControlAction)
			log.Fatal(err)
		}
		return
	}

	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
