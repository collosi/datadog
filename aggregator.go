package datadog

import (
	ddjson "github.com/laslowh/datadog/json"
	"sync"
	"time"
)

type Count struct {
	lock    sync.RWMutex
	series  ddjson.Series
	client  *Client
	updated bool
}

func (c *Client) NewCount(metric string, host string, device string, tags []string) *Count {
	cnt := &Count{
		series: ddjson.Series{
			Metric: metric,
			Points: ddjson.NewCountPointArray(1),
			Type:   "counter",
			Host:   host,
			Device: device,
			Tags:   tags,
		},
		client: c,
	}
	c.counts = append(c.counts, cnt)
	return cnt
}

func (ca *Count) Add(u uint64) {
	t := time.Now().Unix()
	timestamp := &ca.series.Points.Timestamps[0]
	count := &ca.series.Points.CountContent[0]

	ca.lock.Lock()
	if t == *timestamp {
		(*count) += u
	} else {
		(*timestamp) = t
		(*count) += u
	}
	ca.updated = true
	ca.lock.Unlock()
}
