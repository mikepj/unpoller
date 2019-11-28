package promunifi

import (
	"github.com/prometheus/client_golang/prometheus"
	"golift.io/unifi"
)

// Each site has five subsystems.
// Some of them share data, but generally each one has its own data set.
const (
	subsystemLAN  = "lan"
	subsystemVPN  = "vpn"
	subsystemWWW  = "www"
	subsystemWLAN = "wlan"
	subsystemWAN  = "wan"
)

type site struct {
	NumUser               *prometheus.Desc
	NumGuest              *prometheus.Desc
	NumIot                *prometheus.Desc
	TxBytesR              *prometheus.Desc
	RxBytesR              *prometheus.Desc
	NumAp                 *prometheus.Desc
	NumAdopted            *prometheus.Desc
	NumDisabled           *prometheus.Desc
	NumDisconnected       *prometheus.Desc
	NumPending            *prometheus.Desc
	NumGw                 *prometheus.Desc
	NumSw                 *prometheus.Desc
	NumSta                *prometheus.Desc
	Latency               *prometheus.Desc
	Drops                 *prometheus.Desc
	Uptime                *prometheus.Desc
	XputUp                *prometheus.Desc
	XputDown              *prometheus.Desc
	SpeedtestPing         *prometheus.Desc
	RemoteUserNumActive   *prometheus.Desc
	RemoteUserNumInactive *prometheus.Desc
	RemoteUserRxBytes     *prometheus.Desc
	RemoteUserTxBytes     *prometheus.Desc
	RemoteUserRxPackets   *prometheus.Desc
	RemoteUserTxPackets   *prometheus.Desc
}

func descSite(ns string) *site {
	if ns += "_site_"; ns == "_site_" {
		ns = "site_"
	}
	labels := []string{"subsystem", "status", "name", "desc", "site_name"}

	return &site{
		NumUser:               prometheus.NewDesc(ns+"num_user", "Number of Users", labels, nil),
		NumGuest:              prometheus.NewDesc(ns+"num_guest", "Number of Guests", labels, nil),
		NumIot:                prometheus.NewDesc(ns+"num_iot", "Number of IoT Devices", labels, nil),
		TxBytesR:              prometheus.NewDesc(ns+"transmit_rate_bytes", "Bytes Transmit Rate", labels, nil),
		RxBytesR:              prometheus.NewDesc(ns+"receive_rate_bytes", "Bytes Receive Rate", labels, nil),
		NumAp:                 prometheus.NewDesc(ns+"num_ap", "Access Point Count", labels, nil),
		NumAdopted:            prometheus.NewDesc(ns+"num_adopted", "Adoption Count", labels, nil),
		NumDisabled:           prometheus.NewDesc(ns+"num_disabled", "Disabled Count", labels, nil),
		NumDisconnected:       prometheus.NewDesc(ns+"num_disconnected", "Disconnected Count", labels, nil),
		NumPending:            prometheus.NewDesc(ns+"num_pending", "Pending Count", labels, nil),
		NumGw:                 prometheus.NewDesc(ns+"num_gateways", "Gateway Count", labels, nil),
		NumSw:                 prometheus.NewDesc(ns+"num_switches", "Switch Count", labels, nil),
		NumSta:                prometheus.NewDesc(ns+"num_stations", "Station Count", labels, nil),
		Latency:               prometheus.NewDesc(ns+"latency_seconds", "Latency", labels, nil),
		Uptime:                prometheus.NewDesc(ns+"uptime_seconds", "Uptime", labels, nil),
		Drops:                 prometheus.NewDesc(ns+"intenet_drops_total", "Internet (WAN) Disconnections", labels, nil),
		XputUp:                prometheus.NewDesc(ns+"xput_up_rate", "Speedtest Upload", labels, nil),
		XputDown:              prometheus.NewDesc(ns+"xput_down_rate", "Speedtest Download", labels, nil),
		SpeedtestPing:         prometheus.NewDesc(ns+"speedtest_ping", "Speedtest Ping", labels, nil),
		RemoteUserNumActive:   prometheus.NewDesc(ns+"num_remote_user_active", "Remote Users Active", labels, nil),
		RemoteUserNumInactive: prometheus.NewDesc(ns+"num_remote_user_inactive", "Remote Users Inactive", labels, nil),
		RemoteUserRxBytes:     prometheus.NewDesc(ns+"remote_user_receive_bytes_total", "Remote Users Receive Bytes", labels, nil),
		RemoteUserTxBytes:     prometheus.NewDesc(ns+"remote_user_transmit_bytes_total", "Remote Users Transmit Bytes", labels, nil),
		RemoteUserRxPackets:   prometheus.NewDesc(ns+"remote_user_receive_packets_total", "Remote Users Receive Packets", labels, nil),
		RemoteUserTxPackets:   prometheus.NewDesc(ns+"remote_user_transmit_packets_total", "Remote Users Transmit Packets", labels, nil),
	}
}

func (u *unifiCollector) exportSites(r *Report) {
	if r.Metrics == nil || len(r.Metrics.Sites) < 1 {
		return
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		for _, s := range r.Metrics.Sites {
			u.exportSite(r, s)
		}
	}()
}

func (u *unifiCollector) exportSite(r *Report, s *unifi.Site) {
	labels := []string{s.Name, s.Desc, s.SiteName}
	for _, h := range s.Health {
		l := append([]string{h.Subsystem, h.Status}, labels...)

		if h.Subsystem != subsystemVPN {
			r.send([]*metricExports{
				{u.Site.TxBytesR, prometheus.GaugeValue, h.TxBytesR, l},
				{u.Site.RxBytesR, prometheus.GaugeValue, h.RxBytesR, l},
			})
		} else {
			r.send([]*metricExports{
				{u.Site.RemoteUserNumActive, prometheus.GaugeValue, h.RemoteUserNumActive, l},
				{u.Site.RemoteUserNumInactive, prometheus.GaugeValue, h.RemoteUserNumInactive, l},
				{u.Site.RemoteUserRxBytes, prometheus.CounterValue, h.RemoteUserRxBytes, l},
				{u.Site.RemoteUserTxBytes, prometheus.CounterValue, h.RemoteUserTxBytes, l},
				{u.Site.RemoteUserRxPackets, prometheus.CounterValue, h.RemoteUserRxPackets, l},
				{u.Site.RemoteUserTxPackets, prometheus.CounterValue, h.RemoteUserTxPackets, l},
			})
		}

		if h.Subsystem == subsystemWWW {
			r.send([]*metricExports{
				{u.Site.Uptime, prometheus.GaugeValue, h.Latency, l},
				{u.Site.Latency, prometheus.GaugeValue, h.Latency.Val / 1000, l},
				{u.Site.XputUp, prometheus.GaugeValue, h.XputUp, l},
				{u.Site.XputDown, prometheus.GaugeValue, h.XputDown, l},
				{u.Site.SpeedtestPing, prometheus.GaugeValue, h.SpeedtestPing, l},
				{u.Site.Drops, prometheus.CounterValue, h.Drops, l},
			})
		}

		if h.Subsystem == subsystemLAN || h.Subsystem == subsystemWLAN || h.Subsystem == subsystemWAN {
			r.send([]*metricExports{
				{u.Site.NumAdopted, prometheus.GaugeValue, h.NumAdopted, l},
				{u.Site.NumDisconnected, prometheus.GaugeValue, h.NumDisconnected, l},
				{u.Site.NumPending, prometheus.GaugeValue, h.NumPending, l},
			})

			if h.Subsystem == subsystemLAN || h.Subsystem == subsystemWLAN {
				r.send([]*metricExports{
					{u.Site.NumUser, prometheus.GaugeValue, h.NumUser, l},
					{u.Site.NumGuest, prometheus.GaugeValue, h.NumGuest, l},
					{u.Site.NumIot, prometheus.GaugeValue, h.NumIot, l},
				})
			}

			if h.Subsystem == subsystemWLAN {
				r.send([]*metricExports{
					{u.Site.NumAp, prometheus.GaugeValue, h.NumAp, l},
					{u.Site.NumDisabled, prometheus.GaugeValue, h.NumDisabled, l},
				})
			}

			if h.Subsystem == subsystemWAN {
				r.send([]*metricExports{
					{u.Site.NumGw, prometheus.GaugeValue, h.NumGw, l},
					{u.Site.NumSta, prometheus.GaugeValue, h.NumSta, l},
				})
			}

			if h.Subsystem == subsystemLAN {
				r.send([]*metricExports{
					{u.Site.NumSw, prometheus.GaugeValue, h.NumSw, l},
				})
			}
		}
	}
}
