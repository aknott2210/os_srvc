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
var consul string
var config string

type program struct{}

func init() {
	flag.StringVar(&consul, "consul", "", "The path to Consul.")
	flag.StringVar(&config, "config", "", "The path to Consul configuration.")
}

func init() {
        if !arguments.ServiceCall() {
           consul = os.Args[1]     
           config = os.Args[2]
        }
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logger.Info("Starting service...")
	go p.run()
	return nil
}

func (p *program) run() {
       cmd := exec.Command(consul, "agent", "-config-dir", config)
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
		Name:        "Consul8",
		DisplayName: "Consul8",
		Description: "This service starts up Consul",
		Arguments: []string{consul, config},
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