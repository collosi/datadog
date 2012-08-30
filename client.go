package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	ddjson "github.com/laslowh/datadog/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	API_HOST          = "https://app.datadoghq.com"
	METRIC_UPDATE_URL = API_HOST + "/api/v1/series"
	JSON_MIME_TYPE    = "application/json"
)

type Client struct {
	APIKey         string
	ApplicationKey string
	authSuffix     string
	closeChan      chan bool
	counts         []*Count
}

func (c *Client) Start() {
	v := make(url.Values)
	v.Add("api_key", c.APIKey)
	if c.ApplicationKey != "" {
		v.Add("application_key", c.ApplicationKey)
	}
	c.authSuffix = "?" + v.Encode()

	c.closeChan = make(chan bool)
	go c.doUpdates()
}

func (c *Client) UpdateMetrics(s []ddjson.Series) error {
	rp, wp := io.Pipe()
	enc := json.NewEncoder(wp)
	go func() {
		enc.Encode(&s)
		wp.Close()
	}()
	return c.updateMetrics(rp)
}

func (c *Client) updateMetrics(rd io.Reader) error {
	url := METRIC_UPDATE_URL + c.authSuffix
	println(url)
	r, err := http.Post(url, JSON_MIME_TYPE, rd)
	if err != nil {
		return err
	}

	r.Write(os.Stderr)
	if r.StatusCode != http.StatusAccepted {
		return fmt.Errorf("%s: unexpected response code", r.Status)
	}
	return nil
}

func (c *Client) doUpdates() {
	tick := time.NewTicker(100 * time.Millisecond)
	println("starting update loop")
	for {
		select {
		case _, ok := <-c.closeChan:
			println("done update loop")
			if !ok {
				return
			}
		case _ = <-tick.C:
			println("tick", len(c.counts))
			buf := bytes.NewBuffer(make([]byte, 0, 1024))
			hasOne := false
			first := true
			io.WriteString(buf, `{"series": [`)
			for _, count := range c.counts {
				println("checking count")
				if count.updated {
					println("updated")
					hasOne = true
					count.lock.Lock()
					count.updated = false
					count.lock.Unlock()

					if !first {
						io.WriteString(buf, ",")
					}
					enc := json.NewEncoder(buf)
					count.lock.RLock()
					enc.Encode(&count.series)
					count.lock.RUnlock()
					first = false
				}
			}
			if hasOne {
				io.WriteString(buf, "]}")
				println("updating", buf.String())
				err := c.updateMetrics(buf)
				if err != nil {
					println(err.Error())
				}
			}
		}
	}
	println("done update function")
}
