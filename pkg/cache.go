package pkg

import (
	"encoding/json"
	"log"
)

type Cache struct {

}

const (
  mockPod string = `
{
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
		"annotations": {
			"prometheus.io/scrape": "false"
		},
		"labels": {
			"name": "sms-inbound",
			"pod-template-hash": "3065056882"
		},
		"name": "sms-inbound-3065056882-xoast",
		"namespace": "mc-red"
	}
}`
)

func (c *Cache) getPods() []Pod {
	var p Pod
	err := json.Unmarshal([]byte(mockPod), &p)
	if err != nil {
		log.Printf("Couldnot unmarshall pod %v\n%v\n", mockPod, err)
	}
	return []Pod{p}
}

