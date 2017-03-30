// Copyright 2015 Daniel Theophanes.
// Use of this source code is governed by a zlib-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"github.com/aknott2210/os_service/arguments"
	"github.com/aknott2210/os_service/http"
	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
	"log"
	"os"
	"path"
	"strconv"
)

var logger service.Logger
var address string
var port int
var job string

type program struct{}

func init() {
	flag.StringVar(&job, "job", "clarify", "The name of the job to run.")
	flag.StringVar(&address, "address", "localhost", "The http address of Nomad.")
	flag.IntVar(&port, "port", 4646, "The port that Nomad is running on.")
}

func init() {
	if !arguments.ServiceCall() {
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
	if jobRunning() {
		logger.Info("Detected job as running...")
		host := host(hostname())
		if host.Drain {
			logger.Info("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=true")
			http.DrainWithRetry(logger, &client.NomadServer{address, port}, host.ID, false, 3, 3)
			logger.Info("Sent request for node drain enable=false")
		}
	} else {
		logger.Info("Detected no running jobs, submitting " + job)
		http.SubmitJobWithRetry(logger, &client.NomadServer{address, port}, exeDir()+"/../../launch_clarify.json", 5, 3)
	}
}

func (p *program) Stop(s service.Service) error {
	logger.Info("Stopping service...")
	host := host(hostname())
	if !host.Drain {
		logger.Info("Detected node: " + host.Name + " with host/node id: " + host.ID + " as having drain enable=false")
		http.DrainWithRetry(logger, &client.NomadServer{address, port}, host.ID, true, 3, 3)
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
	jobs := http.JobsWithRetry(logger, &client.NomadServer{address, port}, 3, 3)
	for _, nomadJob := range jobs {
		if job == nomadJob.Name {
			return true
		}
	}
	return false
}

func host(hostname string) *client.Host {
	hosts := http.HostsWithRetry(logger, &client.NomadServer{address, port}, 3, 3)
	for _, host := range hosts {
		if hostname == host.Name {
			return &host
		}
	}
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
		Arguments:   []string{address, strconv.Itoa(port), job},
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
