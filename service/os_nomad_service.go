// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// simple does nothing except block while running the service.
package main

import (
	"log"
	"fmt"
	"os"
	"flag"
	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
)

var logger service.Logger
var nomad string
var jobName string

func init() {
	flag.StringVar(&nomad, "nomad", "", "Host address and port of nomad server")
	flag.StringVar(&jobName, "job", "", "Job Name")
}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	fmt.Println("Starting")
	go p.run()
	return nil
}

func (p *program) run() {
       fmt.Println("Running")
       if(jobRunning()) {
           fmt.Println("Job is running")
           if(drained()) {
               fmt.Println("Node has been drained")
           } else {
               fmt.Println("Node has NOT been drained")
           }
           //TODO see if status is drained
       } else {
           //TODO submit job
       }
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	fmt.Println("Stopping")
	return nil
}

func jobRunning() bool {
       jobs := client.PopulateJobs("http://" + nomad)
       for _, job := range jobs {
       		if(jobName == job.Name) {
       		    return true
       		}
       }
       return false
}

func drained() bool {
    hosts := client.PopulateHosts("http://" + nomad)
    hostname, err := os.Hostname() 
    if err != nil {
            log.Fatal(err)
    }
    for _, host := range hosts {
           if(hostname == host.Name && host.Drain) {
           	return true
           }
    }
    return false
}

func main() {
        svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	if nomad == "" {
			fmt.Println("nomad flag must be set.")
			os.Exit(-1)
	}
	if jobName == "" {
	                fmt.Println("job flag must be set.")
			os.Exit(-1)
	}
	
	
	svcConfig := &service.Config{
		Name:        "GoServiceExampleSimple",
		DisplayName: "Go Service Example",
		Description: "This is an example Go service.",
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