/*
Copyright (c) 2014 Ashley Jeffs

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, sub to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package leapnet

import (
	"errors"
	"github.com/jeffail/leaps/leaplib"
	"net/http"
	"time"
)

/*--------------------------------------------------------------------------------------------------
 */

/*
StatsServerConfig - Holds configuration options for the StatsServer
*/
type StatsServerConfig struct {
	Path           string `json:"path"`
	Address        string `json:"address"`
	StaticFilePath string `json:"www_path"`
	StatsTimeout   int    `json:"stat_timeout_ms"`
	RequestTimeout int    `json:"request_timeout_s"`
}

/*
DefaultStatsServerConfig - Returns a fully defined StatsServer configuration with the default values
for each field.
*/
func DefaultStatsServerConfig() StatsServerConfig {
	return StatsServerConfig{
		Path:           "/leapstats",
		Address:        ":4040",
		StaticFilePath: "",
		StatsTimeout:   200,
		RequestTimeout: 10,
	}
}

/*--------------------------------------------------------------------------------------------------
 */

/*
StatsServer - A server constructed to present an HTTP endpoint for obtaining live statics regarding
the leaps server. Requires a reference to the logger shared with the Curator object at the center
of the service.
*/
type StatsServer struct {
	config   StatsServerConfig
	logger   *leaplib.LeapsLogger
	server   *http.Server
	serveMux *http.ServeMux
}

/*
CreateStatsServer - Create a new leaps StatsServer.
*/
func CreateStatsServer(logger *leaplib.LeapsLogger, config StatsServerConfig) (*StatsServer, error) {
	statsServer := StatsServer{
		config:   config,
		logger:   logger,
		server:   nil,
		serveMux: http.NewServeMux(),
	}
	if len(statsServer.config.Address) == 0 || len(statsServer.config.Path) == 0 {
		return nil, errors.New("invalid config value for Address/Path")
	}
	if len(statsServer.config.StaticFilePath) > 0 {
		statsServer.serveMux.Handle("/",
			http.StripPrefix("/", http.FileServer(http.Dir(statsServer.config.StaticFilePath))))
	}
	statsServer.server = &http.Server{
		Addr:           statsServer.config.Address,
		Handler:        &statsServer,
		ReadTimeout:    time.Duration(statsServer.config.RequestTimeout) * time.Second,
		WriteTimeout:   time.Duration(statsServer.config.RequestTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	return &statsServer, nil
}

/*--------------------------------------------------------------------------------------------------
 */

/*
StatsHandler - The StatsServer request handler.
*/
func (s *StatsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.config.Path == r.URL.Path {
		if stats, err := s.logger.GetStats(time.Duration(s.config.StatsTimeout) * time.Millisecond); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(*stats))
		}
	} else {
		s.serveMux.ServeHTTP(w, r)
	}
}

/*
Listen - Bind to the configured http endpoint and begin serving requests.
*/
func (s *StatsServer) Listen() error {
	if len(s.config.Address) == 0 {
		return errors.New("invalid config value for Address")
	}
	err := s.server.ListenAndServe()
	return err
}

/*--------------------------------------------------------------------------------------------------
 */