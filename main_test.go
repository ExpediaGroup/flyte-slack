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
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/flyte-slack/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPackDefinitionIsPopulated(t *testing.T) {

	packDef := GetPackDef(DummySlack{})

	assert.Equal(t, "Slack", packDef.Name)
	assert.Equal(t, "https://github.com/HotelsDotCom/flyte-slack/blob/master/README.md", packDef.HelpURL.String())
	require.Equal(t, 0, len(packDef.Labels))
	require.Equal(t, 2, len(packDef.Commands))
	require.Equal(t, 1, len(packDef.EventDefs))
}

// --- dummy Slack implementation ---

type DummySlack struct{}

func (DummySlack) SendMessage(message, channelId, threadTimestamp string) {}

func (DummySlack) SendRichMessage(client.RichMessage) error { return nil }

func (DummySlack) IncomingMessages() <-chan flyte.Event {
	return make(chan flyte.Event)
}
