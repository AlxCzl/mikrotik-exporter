package collector

import (
	"strings"
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
	c.props = []string{"state"}
	// Description for the returned values in the metric
	labelNames := []string{"devicename", "interface"}
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
		for _, prop := range c.props {
			v, ok := e.Map[prop]
			if !ok {
				continue
			}

			value := float64(c.valueForProp(prop, v))
			ctx.ch <- prometheus.MustNewConstMetric(c.descriptions[prop], prometheus.GaugeValue, value, ctx.device.Name, e.Map["name"])
		}
	}

	return nil
}

// Here are the corresponding values for detect-internet's output:
// - internet: 2
// - wan: 1
// - lan: 0
func (c *internetCollector) valueForProp(name, value string) int {
	switch {
	case name == "state":
		return func(v string) int {
			if v == "internet" {
				return 2
			}
			if v == "wan" {
				return 1
			}
			return 0
		}(value)
	default:
		return 0
	}
}
