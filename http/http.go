package http

import (
	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
	"time"
)

const http_ok_status string = "200 OK"

type post interface {
	request() (string, error)
	toVerb() string
}

type get interface {
	request(target *interface{}) (string, error)
	structType() string
}

type Drain struct {
	nomad  *client.NomadServer
	id     string
	enable bool
	verb   string
}

type SubmitJob struct {
	nomad          *client.NomadServer
	launchFilePath string
	verb           string
}

type Hosts struct {
	nomad *client.NomadServer
}

type Jobs struct {
	nomad *client.NomadServer
}

func (hosts Hosts) request(target *interface{}) (string, error) {
	var err error
	var status string
	*target, status, err = client.Hosts(hosts.nomad)
	return status, err
}

func (jobs Jobs) request(target *interface{}) (string, error) {
	var err error
	var status string
	*target, status, err = client.Jobs(jobs.nomad)
	return status, err
}

func (hosts Hosts) structType() string {
	return "Hosts"
}

func (jobs Jobs) structType() string {
	return "Jobs"
}

func (drain Drain) request() (string, error) {
	return client.Drain(drain.nomad, drain.id, drain.enable)
}

func (submitJob SubmitJob) request() (string, error) {
	return client.SubmitJob(submitJob.nomad, submitJob.launchFilePath)
}

func (drain Drain) toVerb() string {
	return drain.verb
}

func (submitJob SubmitJob) toVerb() string {
	return submitJob.verb
}

func executeHTTPPostWithRetry(logger service.Logger, request post, retries int, sleepSeconds time.Duration) {
	numRetries := 1
	for numRetries < retries {
		status, err := request.request()
		if status == http_ok_status && err == nil {
			logger.Info("Successful http put for: " + request.toVerb() + " with status: " + status)
			return
		}
		numRetries++
		time.Sleep((sleepSeconds * 1000) * time.Millisecond)
		logger.Error("Error "+request.toVerb()+" got http status: "+status+" with error: ", err)
		logger.Error("Retrying...")
	}
	panic("Exceeded max number of retries for " + request.toVerb())
}

func executeHTTPGetWithRetry(logger service.Logger, target *interface{}, request get, retries int, sleepSeconds time.Duration) {
	numRetries := 1
	for numRetries < retries {
		status, err := request.request(target)
		if status == http_ok_status && err == nil {
			logger.Info("Successful http get for: " + request.structType() + " with status: " + status)
			return
		}
		numRetries++
		time.Sleep((sleepSeconds * 1000) * time.Millisecond)
		logger.Error("Error requesting: "+request.structType()+" with error: ", err)
		logger.Error("Retrying...")
	}
	panic("Exceeded max number of retries for get: " + request.structType())
}

func HostsWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Host {
	var hosts interface{}
	executeHTTPGetWithRetry(logger, &hosts, Hosts{nomad}, retries, sleepSeconds)
	return hosts.([]client.Host)
}

func JobsWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Job {
	var jobs interface{}
	executeHTTPGetWithRetry(logger, &jobs, Hosts{nomad}, retries, sleepSeconds)
	return jobs.([]client.Job)
}

func SubmitJobWithRetry(logger service.Logger, nomad *client.NomadServer, launchFilePath string, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, SubmitJob{nomad, launchFilePath, "submitting job"}, retries, sleepSeconds)
}

func DrainWithRetry(logger service.Logger, nomad *client.NomadServer, id string, enable bool, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, Drain{nomad, id, enable, "draining"}, retries, sleepSeconds)
}
