/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"net/url"
	"os"
	"github.com/HotelsDotCom/go-logger"
)

const (
	apiEnvKey            = "FLYTE_API"
	tokenEnvKey          = "FLYTE_SLACK_TOKEN"
	defaultChannelEnvKey = "FLYTE_SLACK_DEFAULT_JOIN_CHANNEL"
	backupDirEnvKey      = "FLYTE_SLACK_BACKUP_DIR"
)

// extracted to variable for testing
var lookupEnv = os.LookupEnv

func ApiHost() *url.URL {

	hostEnv := getEnv(apiEnvKey, true)
	host, err := url.Parse(hostEnv)
	if err != nil {
		logger.Fatalf("%s=%q is not valid URL: %v", apiEnvKey, hostEnv, err)
	}
	return host
}

func SlackToken() string {
	return getEnv(tokenEnvKey, true)
}

func DefaultChannel() string {
	return getEnv(defaultChannelEnvKey, false)
}

func BackupDir() string {
	return getEnv(backupDirEnvKey, false)
}

func getEnv(key string, required bool) string {

	if v, _ := lookupEnv(key); v != "" {
		return v
	}

	if required {
		logger.Fatalf("env=%s not set", key)
	}
	return ""
}
