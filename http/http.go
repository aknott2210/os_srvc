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

type Host struct {
	nomad *client.NomadServer
}

func (host Host) request(target *interface{}) (string, error) {
	var err error
	var status string
	*target, status, err = client.Hosts(host.nomad)
	return status, err
}

func (host Host) structType() string {
	return "Host"
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
		retry := status != http_ok_status || err != nil
		if !retry {
			logger.Info(status)
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
		retry := status != http_ok_status || err != nil
		if !retry {
			logger.Info(status)
			return
		}
		numRetries++
		time.Sleep((sleepSeconds * 1000) * time.Millisecond)
		logger.Error("Error requesting: "+request.structType()+" with error: ", err)
		logger.Error("Retrying...")
	}
	panic("Exceeded max number of retries for get: " + request.structType())
}

func HostWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Host {
	var hosts interface{}
	executeHTTPGetWithRetry(logger, &hosts, Host{nomad}, 3, 5)
	return hosts.([]client.Host)
}

func SubmitJobWithRetry(logger service.Logger, nomad *client.NomadServer, launchFilePath string, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, SubmitJob{nomad, launchFilePath, "submitting job"}, retries, sleepSeconds)
}

func DrainWithRetry(logger service.Logger, nomad *client.NomadServer, id string, enable bool, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, Drain{nomad, id, enable, "draining"}, retries, sleepSeconds)
}
