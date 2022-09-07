/*
Copyright © 2022 Loïc Yavercovski

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package exporter

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var labelNames = []string{
	"computerName",
	"Datacenter",
	"nodeClass",
}

func getClient() *http.Client {
	client := &http.Client{Timeout: 10 * time.Second,
		Transport: &http.Transport{
			MaxConnsPerHost: 1,
		}}
	if !viper.GetBool("disable-authentication") {
		// Load client cert
		cert, err := tls.LoadX509KeyPair(viper.GetString("cert"), viper.GetString("key"))
		if err != nil {
			log.Fatal(err)
		}

		// Load CA cert
		caCert, err := os.ReadFile(viper.GetString("ca"))
		if err != nil {
			log.Fatal(err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			RootCAs:      caCertPool,
		}

		transport := &http.Transport{
			TLSClientConfig: tlsConfig,
			MaxConnsPerHost: 1,
		}
		client = &http.Client{Transport: transport, Timeout: 10 * time.Second}
	}
	return client
}

func getData(client *http.Client, endpoint string) string {
	var url string = viper.GetString("address")
	req, err := http.NewRequest("GET", url+endpoint, nil)
	if err != nil {
		log.Println(err)
		return "unknown"
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return "unknown"
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			return "unknown"
		}
		return string(body)
	}
	log.Printf("getData: failed to get nomad nodes data => %s\n", resp.Status)
	return "unknown"
}

func promWatchNomadNodes(registry prometheus.Registry) {
	maintenanceState := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nomad_node_maintenance_status",
		Help: "The maintenance status of a nomad node 0:ONLINE 1:MAINTENANCE 2:OFFLINE",
	}, labelNames)
	busyState := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nomad_node_busy_status",
		Help: "The busy status of a nomad node 0:IDLE 1:BUSY",
	}, labelNames)
	failGetData := promauto.NewCounter(prometheus.CounterOpts{
		Name: "nomad_node_exporter_failure",
		Help: "The number of failure to get/parse api data since startup",
	})
	registry.MustRegister(maintenanceState)
	registry.MustRegister(busyState)
	registry.MustRegister(failGetData)

	listOfKnownComputer := map[string]computer{}

	go func() {
		client := getClient()
		for {
			// Get Nodes Data
			nodesApiJsonData := getData(client, "/v1/nodes")
			var computerList []computer
			if err := json.Unmarshal([]byte(nodesApiJsonData), &computerList); err != nil {
				log.Printf("promWatchNomadNodes: error parsing nodes JSON (%s) => retrying in 5s\n", err.Error())
				failGetData.Inc()
				time.Sleep(5 * time.Second)
				continue
			}

			// Get Allocations Data
			baseFilter := url.QueryEscape("ClientStatus contains \"running\" and " + viper.GetString("filter"))
			allocationsApiJsonData := getData(client, "/v1/allocations?task_states=False&filter="+baseFilter)
			var allocationList []allocation
			if err := json.Unmarshal([]byte(allocationsApiJsonData), &allocationList); err != nil {
				log.Printf("promWatchNomadNodes: error parsing allocations JSON (%s) => retrying in 5s\n", err.Error())
				failGetData.Inc()
				time.Sleep(5 * time.Second)
				continue
			}

			for _, computer := range computerList {
				_, found := listOfKnownComputer[computer.GetLabelValuesString()]
				if !found {
					listOfKnownComputer[computer.GetLabelValuesString()] = computer
				}

				// If we detect 1 allocation matching the NodeName, it means it's busy running it
				for _, alloc := range allocationList {
					if alloc.NodeName == computer.Name {
						computer.SetBusyStatus(1)
						break
					}
					computer.SetBusyStatus(0)
				}

				maintenanceState.WithLabelValues(computer.GetLabelValues()...).Set(computer.GetMaintenanceStatus())
				busyState.WithLabelValues(computer.GetLabelValues()...).Set(computer.GetBusyStatus())
			}
		L:
			for kname, k := range listOfKnownComputer {
				for _, computer := range computerList {
					if computer.GetLabelValuesString() == kname {
						continue L
					}
				}
				log.Printf("computer %v was removed from master, removing metric..", kname)
				maintenanceState.DeleteLabelValues(k.GetLabelValues()...)
				busyState.DeleteLabelValues(k.GetLabelValues()...)
				delete(listOfKnownComputer, kname)
			}

			time.Sleep(time.Duration(viper.GetDuration("fetch-interval")))
		}
	}()
}

func Entrypoint() {
	r := prometheus.NewRegistry()
	promWatchNomadNodes(*r)
	handler := promhttp.HandlerFor(r, promhttp.HandlerOpts{})

	http.Handle("/metrics", handler)
	addr := fmt.Sprintf(":%d", viper.GetInt("port"))
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Printf("Failed to start Http Listener (%s)\n", err.Error())
	}
}
