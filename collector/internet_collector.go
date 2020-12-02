package collector

import (
	"time"
	"strings"

	"gopkg.in/routeros.v2/proto"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type internetCollector struct {
	props        []string
	descriptions map[string]*prometheus.Desc
}

func newInternetCollector() routerOSCollector {
	c := &internetCollector{}
	c.init()
	return c
}

func (c *internetCollector) init() {
	// Properties other than 'name' that will be used in .proplist
	c.props = []string{"state", "state-change-time"}
	// Description for the returned values in the metric
	labelNames := []string{"devicename", "interface", "state"}
	c.descriptions = make(map[string]*prometheus.Desc)
	for _, p := range c.props {
		c.descriptions[p] = descriptionForPropertyName("internet", p, labelNames)
	}
}

func (c *internetCollector) describe(ch chan<- *prometheus.Desc) {
	for _, d := range c.descriptions {
		ch <- d
	}
}

func (c *internetCollector) collect(ctx *collectorContext) error {
	reply, err := ctx.client.Run("/interface/detect-internet/state/print", "=.proplist=name,"+strings.Join(c.props, ","))
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error fetching detect-internet state")
		return err
	}


	for _, e := range reply.Re {
		c.collectStatusForEth(e.Map["name"], e, ctx)

	}

	return nil
}

// This function parses the status and the last change time of a specific interface
func (c *internetCollector) collectStatusForEth(name string, se *proto.Sentence, ctx *collectorContext) {
	layout := "Jan/02/2006 15:04:05"
	// Parse date
	v, ok := se.Map["state-change-time"]
	if !ok {
		return
	}
	t, err := time.Parse(layout, v)
	if err != nil {
		log.WithFields(log.Fields{
			"device": ctx.device.Name,
			"error":  err,
		}).Error("error parsing detect-internet last state date")
		return
	}

	value := time.Since(t).Seconds()
	ctx.ch <- prometheus.MustNewConstMetric(c.descriptions["state-change-time"], prometheus.GaugeValue, value, ctx.device.Name, se.Map["name"], se.Map["state"])
}


