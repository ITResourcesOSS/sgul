// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// config.go defines commons for application configuration.
package sgul

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

type (
	// Service is the structure for the service information configuration.
	Service struct {
		Group   string
		Name    string
		Version string
	}

	// DB is the structure for the main database configuration.
	DB struct {
		Type       string
		Host       string
		Port       int
		User       string
		Password   string
		Database   string
		Log        bool
		Migrations struct {
			Enabled            bool
			Drop               bool
			SingularTableNames bool
		}
	}

	// Management is the structure for the management http endpoint configuration.
	Management struct {
		Endpoint struct {
			Port            int
			BaseRoutingPath string
		}
		Health struct {
			Path string
			Full bool
		}
	}

	// Log is the structure for the logger configuration.
	// If not present, the Machinery will use a default logger provided
	// by the "gm-log" package.
	Log struct {
		Path       string
		Filename   string
		Console    bool
		Level      string
		JSON       bool
		MaxSize    int
		MaxBackups int
		MaxAge     int
		Compress   bool
		Caller     bool
	}

	// API is the structure for the Http API server and app configuration.
	API struct {
		Endpoint struct {
			Port            int
			BaseRoutingPath string
		}
		Security struct {
			Enabled bool
			Jwt     struct {
				Secret     string
				Expiration struct {
					Enabled bool
					Minutes int32
				}
			}
		}
	}

	// Client defines configuration structure for Http (API) clients.
	Client struct {
		// Timeout specifies a time limit for requests made by this
		// Client. The timeout includes connection time, any
		// redirects, and reading the response body. The timer remains
		// running after Get, Head, Post, or Do return and will
		// interrupt reading of the Response.Body.
		// A Timeout of zero means no timeout.
		Timeout int

		// Timeout is the maximum amount of time a dial will wait for
		// a connect to complete. If Deadline is also set, it may fail
		// earlier.
		// The default is no timeout.
		DialerTimeout int

		// TLSHandshakeTimeout specifies the maximum amount of time waiting to
		// wait for a TLS handshake. Zero means no timeout.
		TLSHandshakeTimeout int

		// ExpectContinueTimeout, if non-zero, specifies the amount of
		// time to wait for a server's first response headers after fully
		// writing the request headers if the request has an
		// "Expect: 100-continue" header. Zero means no timeout and
		// causes the body to be sent immediately, without
		// waiting for the server to approve.
		// This time does not include the time to send the request header.
		ExpectContinueTimeout int

		// ResponseHeaderTimeout, if non-zero, specifies the amount of
		// time to wait for a server's response headers after fully
		// writing the request (including its body, if any). This
		// time does not include the time to read the response body.
		ResponseHeaderTimeout int

		// ServiceRegistry is the endpoint configuration for the
		// service registry used for service discovery by this client.
		// Actually only Type="sgulreg" is managed.
		ServiceRegistry struct {
			Type string
			URL  string
		}
	}

	// Ldap configuration
	Ldap struct {
		Base   string
		Host   string
		Port   int
		UseSSL bool
		Bind   struct {
			DN       string
			Password string
		}
		UserFilter  string
		GroupFilter string
		Attributes  []string
	}

	// AMQP configuration
	AMQP struct {
		User     string
		Password string
		Host     string
		Port     int
		VHost    string
		Exchange string
		Queue    string
	}

	// Configuration describe the type for the configuration file
	Configuration struct {
		Service    Service
		API        API
		DB         DB
		Management Management
		Log        Log
		Ldap       Ldap
		AMQP       AMQP
	}
)

var instance *Configuration
var onceConfig sync.Once

// GetConfiguration returns the Configuration structure singleton instance.
func GetConfiguration() *Configuration {
	onceConfig.Do(func() {
		loadConfiguration()
	})

	return instance
}

func loadConfiguration() {
	viper.SetDefault("logPath", "./log")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if os.Getenv("ENV") != "" {
		viper.SetConfigName("config-" + os.Getenv("ENV"))
	} else {
		viper.SetConfigName("config")
	}

	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	if err := viper.Unmarshal(&instance); err != nil {
		panic(fmt.Errorf("fatal error decoding configuration into struct: %v", err))
	}

}

// LoadConfiguration reads the configuration file according to the ENV var
// and return unmarshalled struct.
func LoadConfiguration(configStruct interface{}) {
	viper.SetDefault("logPath", "./log")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	if os.Getenv("ENV") != "" {
		viper.SetConfigName("config-" + os.Getenv("ENV"))
	} else {
		viper.SetConfigName("config")
	}

	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	if err := viper.Unmarshal(&configStruct); err != nil {
		panic(fmt.Errorf("fatal error decoding configuration into struct: %v", err))
	}
}

// Get returns a configuration map by key. Used for custom or gear configurations.
func Get(key string) interface{} {
	// just in case!
	conf := GetConfiguration()
	if conf == nil {
		panic("No configuration at all!")
	}
	return viper.Get(key)
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key.
func IsSet(key string) bool {
	// just in case!
	conf := GetConfiguration()
	if conf == nil {
		panic("No configuration at all!")
	}
	return viper.IsSet(key)
}
