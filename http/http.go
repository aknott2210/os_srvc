package http

import (
	"net/http"
	"time"

	"github.com/kardianos/service"
	"github.com/pgombola/gomad/client"
)

type post interface {
	request() (int, error)
	toVerb() string
}

type get interface {
	request(target *interface{}) (int, error)
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

func (hosts Hosts) request(target *interface{}) (int, error) {
	_, status, err := client.Hosts(hosts.nomad)
	return status, err
}

func (jobs Jobs) request(target *interface{}) (int, error) {
	_, status, err := client.Jobs(jobs.nomad)
	return status, err
}

func (hosts Hosts) structType() string {
	return "Hosts"
}

func (jobs Jobs) structType() string {
	return "Jobs"
}

func (drain Drain) request() (int, error) {
	return client.Drain(drain.nomad, drain.id, drain.enable)
}

func (submitJob SubmitJob) request() (int, error) {
	return client.SubmitJob(submitJob.nomad, submitJob.launchFilePath)
}

func (drain Drain) toVerb() string {
	return drain.verb
}

func (submitJob SubmitJob) toVerb() string {
	return submitJob.verb
}

func logAndPanic(logger service.Logger, msg string) {
	logger.Error(msg)
	panic(msg)
}

func executeHTTPPostWithRetry(logger service.Logger, request post, retries int, sleepSeconds time.Duration) {
	numRetries := 0
	for numRetries < retries {
		status, err := request.request()
		if status == http.StatusOK && err == nil {
			logger.Infof("Successful http post for: %s with status: %v", request.toVerb(), status)
			return
		}
		numRetries++
		time.Sleep((sleepSeconds * 1000) * time.Millisecond)
		logger.Errorf("Error %s returned %v HTTP status with error: %v", request.toVerb(), status, err)
		logger.Error("Retrying...")
	}

	logAndPanic(logger, "Exceeded max number of retries for "+request.toVerb())
}

func executeHTTPGetWithRetry(logger service.Logger, target *interface{}, request get, retries int, sleepSeconds time.Duration) {
	numRetries := 0
	for numRetries < retries {
		status, err := request.request(target)
		if status == http.StatusOK && err == nil {
			logger.Infof("Successful HTTP get for: %s with status: %v", request.structType(), status)
			return
		}
		numRetries++
		time.Sleep((sleepSeconds * 1000) * time.Millisecond)
		logger.Error("Error requesting: "+request.structType()+" with error: ", err)
		logger.Error("Retrying...")
	}
	logAndPanic(logger, "Exceeded max number of retries for get: "+request.structType())
}

func HostsWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Host {
	var hosts interface{}
	executeHTTPGetWithRetry(logger, &hosts, Hosts{nomad}, retries, sleepSeconds)
	return hosts.([]client.Host)
}

func JobsWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Job {
	var jobs interface{}
	executeHTTPGetWithRetry(logger, &jobs, Jobs{nomad}, retries, sleepSeconds)
	return jobs.([]client.Job)
}

func SubmitJobWithRetry(logger service.Logger, nomad *client.NomadServer, launchFilePath string, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, SubmitJob{nomad, launchFilePath, "submitting job"}, retries, sleepSeconds)
}

func DrainWithRetry(logger service.Logger, nomad *client.NomadServer, id string, enable bool, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, Drain{nomad, id, enable, "draining"}, retries, sleepSeconds)
}
