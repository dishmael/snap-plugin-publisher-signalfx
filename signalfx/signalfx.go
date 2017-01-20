/*
 * http://www.apache.org/licenses/LICENSE-2.0.txt
 *
 * Copyright 2017 OpsVision Solutions
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * 	http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package signalfx

import (
	"bytes"
	"fmt"
	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/signalfx/golib/datapoint"
	"github.com/signalfx/golib/sfxclient"
	"golang.org/x/net/context"
	"log"
	"os"
	"reflect"
	"strings"
)

const (
	NS_VENDOR = "opsvision"
	NS_PLUGIN = "signalfx"
	VERSION   = 1
)

var fileHandle *os.File

type SignalFx struct {
	initialized bool
	token       string
	hostname    string
	namespace   string
}

// Constructor
func New() *SignalFx {
	return new(SignalFx)
}

func (s *SignalFx) init() error {
	s.initialized = true

	return nil
}

/**
 * Returns the configPolicy for the plugin
 */
func (s *SignalFx) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	// The SignalFx token
	policy.AddNewStringRule([]string{NS_VENDOR, NS_PLUGIN},
		"token",
		true)

	// The hostname to use (defaults to local hostname)
	policy.AddNewStringRule([]string{NS_VENDOR, NS_PLUGIN},
		"hostname",
		false)

	return *policy, nil
}

/**
 * Publish metrics to SignalFx using the TOKEN found in the config
 */
func (s *SignalFx) Publish(mts []plugin.Metric, cfg plugin.Config) error {
	// Make sure we've initialized
	if !s.initialized {
		s.init()
	}

	// Set the output file
	f, err := os.OpenFile("/tmp/signalfx-plugin.debug", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer f.Close()

	log.SetOutput(f)
	log.Printf("Inside publisher")

	// Fetch the token
	token, err := cfg.GetString("token")
	if err != nil {
		return err
	}
	s.token = token

	// Attempt to set the hostname
	hostname, err := cfg.GetString("hostname")
	if err != nil {
		hostname, err = os.Hostname()
		if err != nil {
			hostname = "localhost"
		}
	}
	s.hostname = hostname

	// Iterate over the supplied metrics
	for _, m := range mts {
		var buffer bytes.Buffer

		// Convert the namespace to dot notation
		fmt.Fprintf(&buffer, "snap.%s", strings.Join(m.Namespace.Strings(), "."))
		s.namespace = buffer.String()

		// Do some type conversion and send the data
		switch v := m.Data.(type) {
		case uint:
			s.sendIntValue(int64(v))
		case uint32:
			s.sendIntValue(int64(v))
		case uint64:
			s.sendIntValue(int64(v))
		case int:
			s.sendIntValue(int64(v))
		case int32:
			s.sendIntValue(int64(v))
		case int64:
			s.sendIntValue(int64(v))
		case float32:
			s.sendFloatValue(float64(v))
		case float64:
			s.sendFloatValue(float64(v))
		default:
			fmt.Printf("Unknown %T: %v\n", v, v)
		}
	}

	return nil
}

/**
 *
 */
func (s *SignalFx) sendIntValue(value int64) {
	client := sfxclient.NewHTTPDatapointSink()
	client.AuthToken = s.token
	ctx := context.Background()
	client.AddDatapoints(ctx, []*datapoint.Datapoint{
		sfxclient.Gauge(s.namespace, map[string]string{
			"host": s.hostname,
		}, value),
	})
}

/**
 *
 */
func (s *SignalFx) sendFloatValue(value float64) {
	client := sfxclient.NewHTTPDatapointSink()
	client.AuthToken = s.token
	ctx := context.Background()
	client.AddDatapoints(ctx, []*datapoint.Datapoint{
		sfxclient.GaugeF(s.namespace, map[string]string{
			"host": s.hostname,
		}, value),
	})
}
