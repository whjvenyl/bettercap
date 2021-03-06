package modules

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bettercap/bettercap/core"
	"github.com/bettercap/bettercap/network"
	"github.com/bettercap/bettercap/session"
)

const eventTimeFormat = "15:04:05"

func (s EventsStream) viewLogEvent(e session.Event) {
	fmt.Printf("[%s] [%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		e.Label(),
		e.Data.(session.LogMessage).Message)
}

func (s EventsStream) viewApEvent(e session.Event) {
	ap := e.Data.(*network.AccessPoint)
	vend := ""
	if ap.Vendor != "" {
		vend = fmt.Sprintf(" (%s)", ap.Vendor)
	}

	if e.Tag == "wifi.ap.new" {
		fmt.Printf("[%s] [%s] WiFi access point %s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			core.Bold(ap.ESSID()),
			core.Green(ap.BSSID()),
			vend)
	} else if e.Tag == "wifi.ap.lost" {
		fmt.Printf("[%s] [%s] WiFi access point %s (%s) lost.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			core.Red(ap.ESSID()),
			ap.BSSID())
	} else {
		fmt.Printf("[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			ap.String())
	}
}

func (s EventsStream) viewEndpointEvent(e session.Event) {
	t := e.Data.(*network.Endpoint)
	vend := ""
	name := ""

	if t.Vendor != "" {
		vend = fmt.Sprintf(" (%s)", t.Vendor)
	}

	if t.Alias != "" {
		name = fmt.Sprintf(" (%s)", t.Alias)
	} else if t.Hostname != "" {
		name = fmt.Sprintf(" (%s)", t.Hostname)
	}

	if e.Tag == "endpoint.new" {
		fmt.Printf("[%s] [%s] Endpoint %s detected as %s%s.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			core.Bold(t.IpAddress),
			core.Green(t.HwAddress),
			vend)
	} else if e.Tag == "endpoint.lost" {
		fmt.Printf("[%s] [%s] Endpoint %s%s lost.\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			core.Red(t.IpAddress),
			name)
	} else {
		fmt.Printf("[%s] [%s] %s\n",
			e.Time.Format(eventTimeFormat),
			core.Green(e.Tag),
			t.String())
	}
}

func (s EventsStream) viewModuleEvent(e session.Event) {
	fmt.Printf("[%s] [%s] %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		e.Data)
}

func (s EventsStream) viewSnifferEvent(e session.Event) {
	se := e.Data.(SnifferEvent)
	misc := ""

	if e.Tag == "net.sniff.leak.http" {
		req := se.Data.(*http.Request)
		if req.Method != "GET" {
			misc += "\n\n"
			misc += fmt.Sprintf("  Method: %s\n", core.Yellow(req.Method))
			misc += fmt.Sprintf("  URL: %s\n", core.Yellow(req.URL.String()))
			misc += fmt.Sprintf("  Headers:\n")
			for name, values := range req.Header {
				misc += fmt.Sprintf("    %s => %s\n", core.Green(name), strings.Join(values, ", "))
			}

			if err := req.ParseForm(); err == nil {
				misc += "  \n  Form:\n\n"
				for key, values := range req.Form {
					misc += fmt.Sprintf("    %s => %s\n", core.Green(key), core.Bold(strings.Join(values, ", ")))
				}
			} else if req.Body != nil {
				b, _ := ioutil.ReadAll(req.Body)
				misc += fmt.Sprintf("  \n  %s:\n\n    %s\n", core.Bold("Body"), string(b))
			}
		}
	} else if se.Data != nil {
		misc = fmt.Sprintf("%s", se.Data)
	}

	fmt.Printf("[%s] [%s] %s %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Message,
		misc)
}

func (s EventsStream) viewSynScanEvent(e session.Event) {
	se := e.Data.(SynScanEvent)
	fmt.Printf("[%s] [%s] Found open port %d for %s\n",
		e.Time.Format(eventTimeFormat),
		core.Green(e.Tag),
		se.Port,
		core.Bold(se.Address))
}

func (s *EventsStream) View(e session.Event, refresh bool) {
	if e.Tag == "sys.log" {
		s.viewLogEvent(e)
	} else if strings.HasPrefix(e.Tag, "endpoint.") {
		s.viewEndpointEvent(e)
	} else if strings.HasPrefix(e.Tag, "wifi.ap.") {
		s.viewApEvent(e)
	} else if strings.HasPrefix(e.Tag, "mod.") {
		s.viewModuleEvent(e)
	} else if strings.HasPrefix(e.Tag, "net.sniff.") {
		s.viewSnifferEvent(e)
	} else if strings.HasPrefix(e.Tag, "syn.scan") {
		s.viewSynScanEvent(e)
	} else {
		fmt.Printf("[%s] [%s] %v\n", e.Time.Format(eventTimeFormat), core.Green(e.Tag), e)
	}

	if refresh {
		s.Session.Refresh()
	}
}
