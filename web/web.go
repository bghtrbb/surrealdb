// Copyright © 2016 Abcum Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//,
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"github.com/abcum/fibre"
	"github.com/abcum/fibre/mw"
	"github.com/abcum/surreal/cnf"
	"github.com/abcum/surreal/log"
)

// Setup sets up the server for remote connections
func Setup(opts *cnf.Options) (err error) {

	s := fibre.Server(opts)

	// Setup routes

	routes(s)

	// Set timeouts

	s.SetWait(5)

	// Set timeouts

	s.SetName("http")

	// Setup errorware

	s.SetHTTPErrorHandler(errors)

	// Set log level

	s.Logger().SetLogger(log.Instance())

	// Setup middleware

	s.Use(conf())    // Setup conf
	s.Use(mw.Logs()) // Log requests
	s.Use(mw.Fail()) // Catch panics
	s.Use(mw.Gzip()) // Gzip responses
	s.Use(mw.Uniq()) // Add uniq headers
	s.Use(mw.Cors()) // Add cors headers

	// Check body size

	s.Use(mw.Size(&mw.SizeOpts{
		AllowedLength: 1000000,
	}))

	// Check body type

	s.Use(mw.Type(&mw.TypeOpts{
		AllowedContent: []string{
			"application/json",
			"application/msgpack",
		},
	}))

	// Setup basic authentication

	s.Use(mw.Auth(&mw.AuthOpts{
		User: []byte(opts.Auth.User),
		Pass: []byte(opts.Auth.Pass),
	}).Path("/rpc", "/sql", "/key"))

	// Setup special authentication

	s.Use(mw.Sign(&mw.SignOpts{
		Key: []byte(opts.Auth.Token),
		Fnc: func(c *fibre.Context, h, d map[string]interface{}) error {
			c.Set("NS", d["ns"])
			c.Set("DB", d["db"])
			return nil
		},
	}).Path("/rpc", "/sql", "/key"))

	// Setup newrelic integration

	s.Use(mw.Newrelic(&mw.NewrelicOpts{
		Name:    []byte("Surreal"),
		License: []byte(opts.Logging.Newrelic),
	}))

	// Run the server

	if len(opts.Cert.Crt.File) == 0 || len(opts.Cert.Key.File) == 0 {
		s.Run(opts.Conn.Http)
	}

	if len(opts.Cert.Crt.File) != 0 && len(opts.Cert.Key.File) != 0 {
		s.Run(opts.Conn.Http, opts.Cert.Crt.File, opts.Cert.Key.File)
	}

	return nil

}

// Exit tears down the server gracefully
func Exit() {

	log.WithField("prefix", "http").Println("Gracefully shutting down http protocol")

}