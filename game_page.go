package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

type GamePage struct {
	RecordID string
}

func (page *GamePage) BuildURL() string {
	return fmt.Sprintf("https://kif-pona.heroz.jp/games/%s", page.RecordID)
}

var reGamePageRecord = regexp.MustCompile(`receiveMove\("([^"]+)"\)`)

func (page *GamePage) FetchRecord() (*Record, error) {
	url := page.BuildURL()
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("bad status code: %d", res.StatusCode)
	}
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	m := reGamePageRecord.FindStringSubmatch(string(content))
	if m == nil {
		return nil, fmt.Errorf("not matched: %s", reGamePageRecord.String())
	}
	eventStrs := strings.Fields(m[1])

	events := make([]Event, 0, len(eventStrs))
	for _, s := range eventStrs {
		e, err := ParseCSAEvent(s)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return &Record{
		Events: events,
	}, nil
}
