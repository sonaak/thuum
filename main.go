package main

import (
	"flag"
	"github.com/caio/go-tdigest"
	"github.com/golang/glog"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type App struct {
	// takes a config
	// takes an input
	// takes workers
	// takes reporter

	// based on the config construct a planner
	// based on input, retrieves a list of steps
	// pass the steps to the worker pool to get a channel of responses
	// reporter aggregates the response as they come in

}

func discard(body io.ReadCloser) error {
	// TODO: turn into something a bit safer
	defer body.Close()

	_, err := io.Copy(ioutil.Discard, body)
	return err
}

func main() {
	flag.Parse()

	client := http.Client{
		Transport: &http.Transport {
			MaxIdleConns: 1000,
			MaxIdleConnsPerHost: 1000,
		},
	}

	tDigest, err := tdigest.New(tdigest.Compression(100))
	if err != nil {
		panic(err)
	}
	var timingSum float64 = 0.00

	baseUrl := "http://google.com"
	count := 10000
	for i := 0; i < count; i ++ {
		did := uuid.Must(uuid.NewV4())
		url := baseUrl + did.String()

		glog.Infof("Request %s", url)
		requestTime := time.Now()
		request, err := http.NewRequest("GET", url, http.NoBody)
		if err != nil {
			glog.Errorf("Error: %v", err)
		}

		resp, err := client.Do(request)
		if err != nil {
			glog.Errorf("Error: %v", err)
		}

		err = discard(resp.Body)
		if err != nil {
			glog.Errorf("Error: %v", err)
		}

		duration := time.Since(requestTime)
		tDigest.Add(duration.Seconds() * 1000)
		timingSum += duration.Seconds()
		glog.Infof("Duration: %s", duration.String())

		// put in a sleep here so not to be a bot
		time.Sleep(1 * time.Millisecond)
	}

	glog.Infof("P99: %.4f", tDigest.Quantile(.99))
	glog.Infof("Average: %.4f", timingSum / float64(count / 1000))
}
