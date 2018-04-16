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
	"github.com/stretchr/testify/assert"
	"os"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"testing"
)

var TestEnv map[string]string

func BeforeConfig() {

	loggertest.Init("DEBUG")
	TestEnv = make(map[string]string)
	lookupEnv = func(k string) (string, bool) {
		v, ok := TestEnv[k]
		return v, ok
	}
}

func AfterConfig() {

	lookupEnv = os.LookupEnv
	loggertest.Reset()
}

func TestApiHostEnv(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	TestEnv["FLYTE_API"] = "http://test_api:8080"

	url := ApiHost()
	assert.Equal(t, "http://test_api:8080", url.String())
}

func TestApiHostEnvNotSet(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	assert.Panics(t, func() { ApiHost() })
}

func TestApiHostEnvInvalidUrl(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	TestEnv["FLYTE_API"] = ":/invalid url"

	loggertest.Init("DEBUG")
	defer loggertest.Reset()

	assert.Panics(t, func() { ApiHost() })
}

func TestSlackTokenEnv(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	TestEnv["FLYTE_SLACK_TOKEN"] = "abc"

	assert.Equal(t, "abc", SlackToken())
}

func TestSlackTokenEnvNotSet(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	assert.Panics(t, func() { SlackToken() })
}

func TestDefaultChannelEnv(t *testing.T) {
	assert.Equal(t, "", DefaultChannel())
}

func TestDefaultChannelEnvSetWhenNotSet(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	TestEnv["FLYTE_SLACK_DEFAULT_JOIN_CHANNEL"] = "abc"

	assert.Equal(t, "abc", DefaultChannel())
}

func TestBackupDirEnvDefault(t *testing.T) {
	assert.Equal(t, "", BackupDir())
}

func TestBackupDirEnv(t *testing.T) {

	BeforeConfig()
	defer AfterConfig()

	TestEnv["FLYTE_SLACK_BACKUP_DIR"] = "/tmp/slack-pack"

	assert.Equal(t, "/tmp/slack-pack", BackupDir())
}
