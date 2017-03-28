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
	"github.com/aknott2210/os_service/arguments"
	"strconv"
	"time"
)

var logger service.Logger
var address string
var port int
var job string
const http_200_ok_status string = "200 OK"

type http interface {
   request() (string, error)
   toVerb() string
}

type Drain struct {
    nomad* client.NomadServer
    id string
    enable bool
    verb string
}

type SubmitJob struct {
    nomad* client.NomadServer
    launchFilePath string
    verb string
}

type program struct{}

func init() {
	flag.StringVar(&job, "job", "clarify", "The name of the job to run.")
	flag.StringVar(&address, "address", "localhost", "The http address of Nomad.")
	flag.IntVar(&port, "port", 4646, "The port that Nomad is running on.")
}

func init() {
        if(!arguments.ServiceCall()) {
                address = os.Args[1]
                var err error
	        port, err = strconv.Atoi(os.Args[2])
	        if err != nil {
	              logger.Error(err)
	        }
                job = os.Args[3]
        }
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
               drainRetryIndefinitely(&client.NomadServer{address, port}, host.ID, false)
               logger.Info("Sent request for node drain enable=false")
           } 
       } else {
           logger.Info("Detected no running jobs, submitting " + job)
           submitJobRetryIndefinitely(&client.NomadServer{address, port}, exeDir() + "/../../launch_clarify.json")
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

func submitJobRetryIndefinitely(nomad *client.NomadServer, launchFilePath string) {
    executeHTTPRequestRetryIndefinitely(SubmitJob{nomad, launchFilePath, "submitting job"})
}

func drainRetryIndefinitely(nomad *client.NomadServer, id string, enable bool) {
    executeHTTPRequestRetryIndefinitely(Drain{nomad, id, enable, "draining"})
}

func (drain Drain) request() (string, error) {
    status, err := client.Drain(drain.nomad, drain.id, drain.enable)
    return status, err
}

func (drain Drain) toVerb() string {
    return drain.verb
}

func (submitJob SubmitJob) toVerb() string {
    return submitJob.verb
}

func (submitJob SubmitJob) request() (string, error) {
    return client.SubmitJob(submitJob.nomad, submitJob.launchFilePath)
}

func executeHTTPRequestRetryIndefinitely(request http) {
   for status, err := request.request(); status != http_200_ok_status || err != nil; {
        logger.Error("Error " + request.toVerb() + " got http status: " + status + " with error: ", err)
        time.Sleep(3000 * time.Millisecond)
    } 
}

func host(hostname string) *client.Host {
   hosts := client.Hosts(&client.NomadServer{address, port})
   for _, host := range hosts {
           if(hostname == host.Name) {
           	return &host
           }
   }
   panic("Couldn't detect host")
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