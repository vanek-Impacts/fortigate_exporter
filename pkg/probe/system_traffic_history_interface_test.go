package probe

import (
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestSystemTrafficHistoryInterface(t *testing.T) {
	c := newFakeClient()
	c.prepare("api/v2/monitor/system/interface/select", "testdata/interface-traffic-history-minimal.jsonnet")
	c.prepare("api/v2/monitor/system/traffic-history/interface?vdom=root&interface=wan1&time_period=hour", "testdata/system-traffic-history-interface-wan1-minimal.jsonnet")
	r := prometheus.NewPedanticRegistry()
	if !testProbe(probeSystemTrafficHistoryInterface, c, r) {
		t.Errorf("probeSystemTrafficHistoryInterface() returned non-success")
	}

	em := `
	# HELP fortigate_interface_traffic_history_last_rx Last received traffic value as reported by system traffic-history interface monitor (time_period=hour)
	# TYPE fortigate_interface_traffic_history_last_rx gauge
	fortigate_interface_traffic_history_last_rx{alias="",name="wan1",parent="",vdom="root"} 456
	# HELP fortigate_interface_traffic_history_last_tx Last transmitted traffic value as reported by system traffic-history interface monitor (time_period=hour)
	# TYPE fortigate_interface_traffic_history_last_tx gauge
	fortigate_interface_traffic_history_last_tx{alias="",name="wan1",parent="",vdom="root"} 123
	`
	if err := testutil.GatherAndCompare(r, strings.NewReader(em)); err != nil {
		t.Fatalf("metric compare: err %v", err)
	}
}


