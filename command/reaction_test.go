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

package command

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var ReactionMockSlack *MockSlack

func TestGetReactionListCommandIsPopulated(t *testing.T) {

	command := GetReactionList(ReactionMockSlack)

	assert.Equal(t, "GetReactionList", command.Name)
	require.Equal(t, 2, len(command.OutputEvents))
	assert.Equal(t, "GetReactionListSuccess", command.OutputEvents[0].Name)
	assert.Equal(t, "GetReactionListFailed", command.OutputEvents[1].Name)
}

func TestGetReactionListReturnsGetReactionListSuccess(t *testing.T) {

	BeforeMessage()
	defer AfterMessage()

	handler := GetReactionList(ReactionMockSlack).Handler
	event := handler([]byte(`{"count": 50 , 
					"message": "" ,
					"threadTimestamp":"1645441176.871569",
					"reactionUser":"U01NAB6ERFB", 
 					"channelId": "C019Q472YPJ",
					"threadTimestamp":"1645441176.871569"
					}`))
	output := event.Payload.(GetReactionListOutput)
	assert.Equal(t, "GetReactionListSuccess", event.EventDef.Name)
	assert.Equal(t, "", output.Message)
	assert.Equal(t, "C019Q472YPJ", output.ChannelId)
}
