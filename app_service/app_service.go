// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"flag"
	"github.com/kardianos/service"
	"github.com/aknott2210/os_service/arguments"
	"os"
	"os/exec"
)

var logger service.Logger
var app string
var config string
var configFlag string
var serviceName string

type program struct{}

func init() {
	flag.StringVar(&app, "app", "", "The path to the application.")
	flag.StringVar(&serviceName, "serviceName", "", "The name of the service.")
	flag.StringVar(&config, "config", "", "The path to the configuration.")
	flag.StringVar(&configFlag, "configFlag", "", "The configuration flag to provide to the application.")
}

func init() {
        if !arguments.ServiceCall() {
           app = os.Args[1]     
           config = os.Args[2]
           configFlag = os.Args[3]
           serviceName = os.Args[4]
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
       if err != nil {
            logger.Error(err)
       }
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
        svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: "This service starts up " + serviceName,
		Arguments: []string{app, config, configFlag, serviceName},
	}

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