// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// simple does nothing except block while running the service.
package main

import (
	"log"
	"os"
	"flag"
	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
)

var logger service.Logger
var address string = "10.10.20.31"
var port int = 4646
var jobName string = "clarify"

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logger.Info("Starting service...")
	go p.run()
	return nil
}

func (p *program) run() {
       if(jobRunning()) {
           logger.Info("Detected job as running...")
           host := host("server-1")
           if(host.Drain) {
               logger.Info("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=true")
               client.Drain(&client.NomadServer{address, port}, host.ID, false)
               logger.Info("Sent request for node drain enable=false")
           } 
       } else {
           logger.Info("Detected no running jobs, submitting " + jobName)
           client.SubmitJob(&client.NomadServer{address, port}, "../launch_clarify.json")
       }
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	logger.Info("Stopping service...")
	host := host("server-1")
	if(host.Drain == false) {
	    logger.Info("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=false")
	    client.Drain(&client.NomadServer{address, port}, host.ID, true)
	    logger.Info("Sent request for node drain enable=true")
	} else {
	    logger.Warning("Unexpectedly detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=true")
	}
	return nil
}

func hostname() string {
    hostname, err := os.Hostname() 
    if err != nil {
        log.Fatal(err)
    }
    return hostname
}

func jobRunning() bool {
       jobs := client.Jobs(&client.NomadServer{address, port})
       for _, job := range jobs {
       		if(jobName == job.Name) {
       		    return true
       		}
       }
       return false
}

func host(hostname string) *client.Host {
   hosts := client.Hosts(&client.NomadServer{address, port})
   for _, host := range hosts {
           if(hostname == host.Name) {
           	return &host
           }
   }
   logger.Error("Couldn't detect host, failing fast")
   os.Exit(1)
   return &client.Host{}
}

func main() {
        svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	
	svcConfig := &service.Config{
		Name:        "GoServiceExampleSimple18",
		DisplayName: "Go Service Example18",
		Description: "This is an example Go service18.",
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