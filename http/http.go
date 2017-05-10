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
	request() (interface{}, int, error)
	structType() string
}

type drain struct {
	nomad  *client.NomadServer
	id     string
	enable bool
	verb   string
}

type submitJob struct {
	nomad          *client.NomadServer
	launchFilePath string
	verb           string
}

type hosts struct {
	nomad *client.NomadServer
}

type jobs struct {
	nomad *client.NomadServer
}

func (hosts hosts) request() (interface{}, int, error) {
	return client.Hosts(hosts.nomad)
}

func (jobs jobs) request() (interface{}, int, error) {
	return client.Jobs(jobs.nomad)
}

func (hosts hosts) structType() string {
	return "Hosts"
}

func (jobs jobs) structType() string {
	return "Jobs"
}

func (drain drain) request() (int, error) {
	return client.Drain(drain.nomad, drain.id, drain.enable)
}

func (submitJob submitJob) request() (int, error) {
	return client.SubmitJob(submitJob.nomad, submitJob.launchFilePath)
}

func (drain drain) toVerb() string {
	return drain.verb
}

func (submitJob submitJob) toVerb() string {
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

func executeHTTPGetWithRetry(logger service.Logger, request get, retries int, sleepSeconds time.Duration) interface{} {
	numRetries := 0
	for numRetries < retries {
		target, status, err := request.request()
		if status == http.StatusOK && err == nil {
			logger.Infof("Successful HTTP get for: %s with status: %v", request.structType(), status)
			return target
		}
		numRetries++
		time.Sleep((sleepSeconds * 1000) * time.Millisecond)
		logger.Error("Error requesting: "+request.structType()+" with error: ", err)
		logger.Error("Retrying...")
	}
	logAndPanic(logger, "Exceeded max number of retries for get: "+request.structType())
	return nil
}

func HostsWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Host {
	hosts := executeHTTPGetWithRetry(logger, hosts{nomad}, retries, sleepSeconds)
	return hosts.([]client.Host)
}

func JobsWithRetry(logger service.Logger, nomad *client.NomadServer, retries int, sleepSeconds time.Duration) []client.Job {
	jobs := executeHTTPGetWithRetry(logger, jobs{nomad}, retries, sleepSeconds)
	return jobs.([]client.Job)
}

func SubmitJobWithRetry(logger service.Logger, nomad *client.NomadServer, launchFilePath string, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, submitJob{nomad, launchFilePath, "submitting job"}, retries, sleepSeconds)
}

func DrainWithRetry(logger service.Logger, nomad *client.NomadServer, id string, enable bool, retries int, sleepSeconds time.Duration) {
	executeHTTPPostWithRetry(logger, drain{nomad, id, enable, "draining"}, retries, sleepSeconds)
}
