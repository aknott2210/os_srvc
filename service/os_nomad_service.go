// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

// simple does nothing except block while running the service.
package main

import (
	"log"
	"os"
	"flag"
	"fmt"
	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
)

var logger service.Logger
var nomad string = "10.10.20.31:4646"
var jobName string = "clarify"

//func init() {
//	flag.StringVar(&nomad, "nomad", "", "Host address and port of nomad server")
//	flag.StringVar(&jobName, "job", "", "Job Name")
//}

type program struct{}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logInfo("Starting service...")
	go p.run()
	return nil
}

func (p *program) run() {
       if(jobRunning()) {
           logInfo("Detected job as running...")
           host := host("server-1")
           if(host.Drain) {
               logInfo("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=true")
               client.Drain("http://" + nomad, host.ID, false)
               logInfo("Sent request for node drain enable=false")
           } 
       } else {
           logInfo("Detected no running jobs, submitting " + jobName)
           //TODO submit job
       }
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	logInfo("Stopping service...")
	host := host("server-1")
	if(host.Drain == false) {
	    logInfo("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=false")
	    client.Drain("http://" + nomad, host.ID, true)
	    logInfo("Sent request for node drain enable=true")
	} else {
	    logWarning("Unexpectedly detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=true")
	}
	return nil
}

func logInfo(msg string) {
    fmt.Println(msg)
    logger.Info(msg)
}

func logWarning(msg string) {
    fmt.Println(msg)
    logger.Warning(msg)
}

func logError(msg string) {
    fmt.Println(msg)
    logger.Error(msg)
}

func hostname() string {
    hostname, err := os.Hostname() 
    if err != nil {
        log.Fatal(err)
    }
    return hostname
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

func host(hostname string) *client.Host {
   hosts := client.PopulateHosts("http://" + nomad)
   for _, host := range hosts {
           if(hostname == host.Name) {
           	return &host
           }
   }
   logError("Couldn't detect host, failing fast")
   os.Exit(1)
   return &client.Host{}
}

func main() {
        svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	//if nomad == "" {
	//		fmt.Println("nomad flag must be set.")
	//		os.Exit(-1)
	//}
	//if jobName == "" {
	//                fmt.Println("job flag must be set.")
	//		os.Exit(-1)
	//}
	
	
	svcConfig := &service.Config{
		Name:        "GoServiceExampleSimple17",
		DisplayName: "Go Service Example17",
		Description: "This is an example Go service17.",
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