// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"
	"flag"
	"path"
	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
	"strconv"
)

var logger service.Logger
var address string
var port int
var job string
const serviceFlag string = "-service"

type program struct{}

func init() {
	flag.StringVar(&job, "job", "clarify", "The name of the job to run.")
	flag.StringVar(&address, "address", "localhost", "The http address of Nomad.")
	flag.IntVar(&port, "port", 4646, "The port that Nomad is running on.")
}

func init() {
        if(!serviceCall()) {
                address = os.Args[1]
                var err error
	        port, err = strconv.Atoi(os.Args[2])
	        if err != nil {
	              logger.Error(err)
	        }
                job = os.Args[3]
        }
}

func serviceCall() bool {
    for _, arg := range os.Args {
        if arg == serviceFlag {
                return true
        }
    }
    return false
}

func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	logger.Info("Starting service...")
	go p.run()
	return nil
}

func (p *program) run() {
       if(jobRunning()) {
           logger.Info("Detected job as running...")
           host := host(hostname())
           if(host.Drain) {
               logger.Info("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=true")
               logger.Info(client.Drain(&client.NomadServer{address, port}, host.ID, false))
               logger.Info("Sent request for node drain enable=false")
           } 
       } else {
           logger.Info("Detected no running jobs, submitting " + job)
           logger.Info(client.SubmitJob(&client.NomadServer{address, port}, exeDir() + "/../../launch_clarify.json"))
       }
}

func (p *program) Stop(s service.Service) error {
	// Stop should not block. Return with a few seconds.
	logger.Info("Stopping service...")
	host := host(hostname())
	if(!host.Drain) {
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
       for _, nomadJob := range jobs {
       		if(job == nomadJob.Name) {
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

func exeDir() string {
    ex, err := os.Executable()
    if err != nil {
        panic(err)
    }
    return path.Dir(ex)
}

func main() {
        svcFlag := flag.String("service", "", "Control the system service.")
	flag.Parse()
	svcConfig := &service.Config{
		Name:        "clarify",
		DisplayName: "clarify",
		Description: "This service starts Clarify by making REST calls to Nomad.",
		Arguments: []string{address, strconv.Itoa(port), job},
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