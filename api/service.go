package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/nanopack/portal/cluster"
	"github.com/nanopack/portal/config"
	"github.com/nanopack/portal/core"
	"github.com/nanopack/portal/core/common"
)

func parseReqService(req *http.Request) (*core.Service, error) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		config.Log.Error(err.Error())
		return nil, BodyReadFail
	}

	var svc core.Service

	if err := json.Unmarshal(b, &svc); err != nil {
		return nil, BadJson
	}

	svc.GenId()
	if svc.Id == "--0" {
		return nil, NoServiceError
	}

	for i := range svc.Servers {
		svc.Servers[i].GenId()
	}

	config.Log.Trace("SERVICE: %+v", svc)
	return &svc, nil
}

// Get information about a service
func getService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	service, err := common.GetService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusNotFound)
		return
	}
	writeBody(rw, req, service, http.StatusOK)
}

// Reset all services
// /services
func putServices(rw http.ResponseWriter, req *http.Request) {
	services := []core.Service{}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&services); err != nil {
		writeError(rw, req, BadJson, http.StatusBadRequest)
		return
	}

	for i := range services {
		services[i].GenId()
		if services[i].Id == "--0" {
			writeError(rw, req, NoServiceError, http.StatusBadRequest)
			return
		}
		for j := range services[i].Servers {
			services[i].Servers[j].GenId()
			if services[i].Servers[j].Id == "-0" {
				writeError(rw, req, NoServerError, http.StatusBadRequest)
				return
			}
		}
	}

	// save to cluster
	err := cluster.SetServices(services)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, services, http.StatusOK)
}

// Create a service
func postService(rw http.ResponseWriter, req *http.Request) {
	// /services
	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	// idempotent additions (don't update service on post)
	if svc, _ := common.GetService(service.Id); svc != nil {
		writeBody(rw, req, service, http.StatusOK)
		return
	}

	// save to cluster
	err = cluster.SetService(service)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Replace a service (by replacing all services)
func putService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")
	// rough sanitization
	if len(strings.Split(svcId, "-")) != 3 {
		writeError(rw, req, NoServiceError, http.StatusBadRequest)
		return
	}

	service, err := parseReqService(req)
	if err != nil {
		writeError(rw, req, err, http.StatusBadRequest)
		return
	}

	services, err := common.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	// update service by id
	for i := range services {
		if services[i].Id == svcId {
			services[i] = *service
			break
		}
	}

	// save to cluster
	if err := cluster.SetServices(services); err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, service, http.StatusOK)
}

// Delete a service
func deleteService(rw http.ResponseWriter, req *http.Request) {
	// /services/{svcId}
	svcId := req.URL.Query().Get(":svcId")

	// remove from cluster
	err := cluster.DeleteService(svcId)
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}

	writeBody(rw, req, apiMsg{"Success"}, http.StatusOK)
}

// List all services
func getServices(rw http.ResponseWriter, req *http.Request) {
	// /services
	svcs, err := common.GetServices()
	if err != nil {
		writeError(rw, req, err, http.StatusInternalServerError)
		return
	}
	writeBody(rw, req, svcs, http.StatusOK)
}
