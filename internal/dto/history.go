package dto

import (
	"fmt"
	"time"

	"github.com/DarknessKiller/pingopher/internal/model"
)

type Histories struct {
	HostURL string   `json:"hostUrl"`
	Results []Result `json:"results"`
}

type Result struct {
	DNS        string    `json:"dns"`
	StatusCode uint16    `json:"statusCode"`
	Latency    string    `json:"latency"`
	Timestamp  time.Time `json:"timestamp"`
	ErrorMsg   string    `json:"errorMsg,omitempty"`
}

func ToHistories(histories []*model.History) Histories {
	if len(histories) == 0 {
		return Histories{}
	}

	n := len(histories)
	results := make([]Result, n)

	for i, history := range histories {
		dnsName := history.DNS.Name
		if dnsName == "" {
			dnsName = "System DNS"
		}

		results[i] = Result{
			DNS:        dnsName,
			StatusCode: history.StatusCode,
			Latency:    fmt.Sprintf("%d ms", history.Latency),
			Timestamp:  history.CreatedAt,
			ErrorMsg:   history.ErrorMessage,
		}
	}

	return Histories{
		HostURL: histories[0].Host.HostURL,
		Results: results,
	}
}
