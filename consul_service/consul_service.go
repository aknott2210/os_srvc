// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"flag"
	"github.com/kardianos/service"
	"os"
	//"os/exec"
)

var logger service.Logger
var consulPath string

type program struct{}

func init() {
	flag.StringVar(&consulPath, "consulPath", "", "The path to Consul.")
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logger.Info("Starting service...")
	go p.run()
	return nil
}

func (p *program) run() {
       logger.Error(os.Args[1])
       //cmd := exec.Command("sleep", "5")
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func main() {
        svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	svcConfig := &service.Config{
		Name:        "Consul Service5",
		DisplayName: "Consul Service5",
		Description: "This service starts up Consul",
		Arguments: []string{"TEST ARG"},
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