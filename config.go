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
	"github.com/ExpediaGroup/flyte-slack/cache"
	"github.com/HotelsDotCom/go-logger"
	"net/url"
	"os"
	"strconv"
	"time"
)

const (
	apiEnvKey                 = "FLYTE_API"
	tokenEnvKey               = "FLYTE_SLACK_TOKEN"
	packNameKey               = "PACK_NAME"
	renewConversationList     = "RENEW_CONVERSATION_LIST"  // how often conversation list is updated  cache (hours)
	slackVerificationTokenKey = "SLACK_VERIFICATION_TOKEN" //how to verify the requests
)

// extracted to variable for testing
var lookupEnv = os.LookupEnv

func apiHost() *url.URL {

	hostEnv := getEnv(apiEnvKey, true)
	host, err := url.Parse(hostEnv)
	if err != nil {
		logger.Fatalf("%s=%q is not valid URL: %v", apiEnvKey, hostEnv, err)
	}
	return host
}

func packName() string {
	return getEnv(packNameKey, false)
}

func slackToken() string {
	return getEnv(tokenEnvKey, true)
}

func slackVerificationToken() string {
	return getEnv(slackVerificationTokenKey, true)
}

func cacheConfig() (*cache.Config, error) {
	rc := getEnvDefault(renewConversationList, "24")

	t, err := strconv.Atoi(rc)
	if err != nil {
		return nil, err
	}

	return &cache.Config{
		RenewConversationListFrequency: time.Duration(t) * time.Hour,
	}, nil
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

func getEnvDefault(key string, def string) string {
	if v, _ := lookupEnv(key); v != "" {
		return v
	}

	return def
}
