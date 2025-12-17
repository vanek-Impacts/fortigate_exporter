package probe

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"

	"github.com/bluecmd/fortigate_exporter/pkg/http"
	"github.com/prometheus/client_golang/prometheus"
)

func probeSystemTrafficHistoryInterface(c http.FortiHTTP, meta *TargetMetadata) ([]prometheus.Metric, bool) {
	var (
		mLastTx = prometheus.NewDesc(
			"fortigate_interface_traffic_history_last_tx",
			"Last transmitted traffic value as reported by system traffic-history interface monitor (time_period=hour)",
			[]string{"vdom", "name", "alias", "parent"}, nil,
		)
		mLastRx = prometheus.NewDesc(
			"fortigate_interface_traffic_history_last_rx",
			"Last received traffic value as reported by system traffic-history interface monitor (time_period=hour)",
			[]string{"vdom", "name", "alias", "parent"}, nil,
		)
	)

	type ifResult struct {
		Name      string
		Alias     string
		Interface string
	}
	type ifResponse struct {
		Results map[string]ifResult
		VDOM    string
	}
	var irs []ifResponse

	if err := c.Get("api/v2/monitor/system/interface/select", "vdom=*&include_vlan=true&include_aggregate=true", &irs); err != nil {
		log.Printf("Error: %v", err)
		return nil, false
	}

	type trafficHistoryInterfaceResponse struct {
		Results json.RawMessage `json:"results"`
		LastTx  *float64        `json:"last_tx"`
		LastRx  *float64        `json:"last_rx"`
	}
	type trafficHistoryInterfaceResults struct {
		LastTx *float64 `json:"last_tx"`
		LastRx *float64 `json:"last_rx"`
	}

	m := []prometheus.Metric{}
	success := true
	for _, v := range irs {
		for _, iface := range v.Results {
			if iface.Name == "" {
				continue
			}
			var thr trafficHistoryInterfaceResponse
			query := "interface=" + url.QueryEscape(iface.Name) + "&time_period=hour"
			if v.VDOM != "" {
				query = "vdom=" + url.QueryEscape(v.VDOM) + "&" + query
			}
			if err := c.Get("api/v2/monitor/system/traffic-history/interface", query, &thr); err != nil {
				if strings.Contains(err.Error(), "Response code was 424") {
					// FortiOS returns 424 for some interface types / states.
					// Skip those interfaces to avoid failing the whole scrape.
					continue
				}
				log.Printf("Error: %v", err)
				success = false
				continue
			}

			tx := thr.LastTx
			rx := thr.LastRx
			if (tx == nil || rx == nil) && len(thr.Results) != 0 {
				var tr trafficHistoryInterfaceResults
				if err := json.Unmarshal(thr.Results, &tr); err != nil {
					log.Printf("Error: %v", err)
					success = false
					continue
				}
				if tx == nil {
					tx = tr.LastTx
				}
				if rx == nil {
					rx = tr.LastRx
				}
			}

			if tx == nil || rx == nil {
				log.Printf("Error: traffic-history interface response missing last_tx/last_rx for interface %q", iface.Name)
				success = false
				continue
			}

			m = append(m, prometheus.MustNewConstMetric(mLastTx, prometheus.GaugeValue, *tx, v.VDOM, iface.Name, iface.Alias, iface.Interface))
			m = append(m, prometheus.MustNewConstMetric(mLastRx, prometheus.GaugeValue, *rx, v.VDOM, iface.Name, iface.Alias, iface.Interface))
		}
	}

	return m, success
}


