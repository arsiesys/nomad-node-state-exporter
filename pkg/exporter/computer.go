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
	"strings"
)

type computer struct {
	Name                  string `json:"Name"`
	SchedulingEligibility string `json:"SchedulingEligibility"`
	Status                string `json:"Status"`
	NodeClass             string `json:"NodeClass"`
	Datacenter            string `json:"Datacenter"`
	BusyStatus            float64
	runningAllocations    int
}

func (c *computer) GetLabelValues() []string {
	return []string{
		c.Name,
		c.Datacenter,
		c.NodeClass,
	}
}

func (c *computer) GetLabelValuesString() string {
	return strings.Join(c.GetLabelValues(), "_")
}

func (c *computer) SetBusyStatus(status float64) float64 {
	c.BusyStatus = status
	return c.BusyStatus
}

func (c *computer) GetBusyStatus() float64 {
	// 0: Idle
	// 1: Busy
	return c.BusyStatus
}

func (c *computer) GetMaintenanceStatus() float64 {
	// O: Online
	// 1: Maintenance
	// 2: Offline
	if c.SchedulingEligibility == "ineligible" {
		return 1
	}
	if c.Status == "down" {
		return 2
	}
	return 0
}
