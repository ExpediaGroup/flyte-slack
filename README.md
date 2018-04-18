# flyte-slack ![Build Status](https://travis-ci.org/HotelsDotCom/flyte-slack.svg?branch=master)

A Slack pack for flyte.

## Build

Pack requires go version min. 1.9 and uses dep to manage dependencies (install dep and run `dep ensure`
before build/test)

- go build `go build`
- go test `go test ./...`
- docker build `docker build -t <name>:<version> .`

## Configuration

The plugin is configured using environment variables:

ENV VAR                          | Default  |  Description                               | Example               
 ------------------------------- |  ------- |  ----------------------------------------- |  ---------------------
FLYTE_API                        | -        | The API endpoint to use                    | http://localhost:8080
FLYTE_SLACK_TOKEN                | -        | The API token to use                       | token_abc
FLYTE_SLACK_DEFAULT_JOIN_CHANNEL | -        | A channel to join by default when launched | 1234
FLYTE_SLACK_BACKUP_DIR           | $TMPDIR  | Directory where to backup joined channels  | /flyte-slack

Example `FLYTE_API=http://localhost:8080 FLYTE_SLACK_TOKEN=token_abc FLYTE_SLACK_DEFAULT_JOIN_CHANNEL=1234 ./flyte-slack`

## Commands

All the events have the same fields as the command input plus error (in case of failed event)

### SendMessage

    {
        "message": "...", // required
        "channelId": "..." // required
    }

Returned events

`MessageSent`

    {
        "message": "...",
        "channelId": "..."
    }

`SendMessageFailed`

    {
        "message": "...",
        "channelId": "..."
        "error": "..."
    }

### SendRichMessage

See the [Slack message formatting API](https://api.slack.com/docs/message-formatting) (and [the example below](#rich_message_example))
for details on what can be included in this. All of the fields (at the time of writing) available on the Slack API are supported here.

Returned events

`RichMessageSent`

The returned event payload is the same as the input.

`SendRichMessageFailed`
```json
{
  "inputMessage": { ... },
  "error": "..."
}
```

### Broadcast

Same as send message, but without channel id. Message will be sent to all the channels that pack has joined.

    {
        "message": "..." // required
    }

Returned events

`BroadcastSent`

    {
        "message": "..."
    }

`BroadcastFailed`

    {
        "message": "...",
        "error": "..."
    }

### JoinChannel

Joins channel, pack will start sending `ReceivedMessage` events if there's new message in the joined channel 

    {
        "channelId": "..." // required
    }

Returned events

`ChannelJoined`

    {
        "channelId": "..."
    }

`JoinChannelFailed`

    {
        "channelId": "...",
        "error": "..."
    }

### LeaveChannel

Leaves Slack channel, channel will not be monitored for incoming messages.

    {
        "channelId": "..." // required
    }

Returned events

`ChannelLeft`

    {
        "channelId": "..."
    }

`LeaveChannelFailed`

    {
        "channelId": "...",
        "error": "..."
    }

## Events 

### ReceivedMessage

    {
        "channelId": "...",
        "user": {                // user that sent the message
            "id": "...",
            "name": "...",       // display name
            "email": "...",
            "title": "...",      // e.g. Principal Systems Engineer
            "firstName": "...",
            "lastName": "..."
        },
        "message": "..."
    }

## Example Flows

The following flow will allow you to type any message into a Slack channel where this pack is
listening and have it echo back whatever you say. E.g.: typing `echo hello world` will echo
"hello world" back to you.

```json
{
  "name": "echo",
  "description": "Echo message back to the sender with a mention, like echo service",
  "steps": [
    {
      "id": "get_message",
      "event": {
        "packName": "Slack",
        "name": "ReceivedMessage"
      },
      "criteria": "{{ Event.Payload.message | match: '^echo\\\\s+' }}",
      "context": {
        "Msg": "{{ Event.Payload.message | slice:'5:'}}",
        "Channel": "{{ Event.Payload.channelId }}",
        "RequestorId": "{{ Event.Payload.user.id }}"
      },
      "command": {
        "packName": "Slack",
        "name": "SendMessage",
        "input": {
          "channelId":"{{ Context.Channel }}",
          "message":"<@{{ Context.RequestorId }}>, you said:\n>>> {{ Context.Msg }}"
        }
      }
    }
  ]
}
```

<a name="rich_message_example"></a>
This flow sends a rich message for a hypothetical deployment. More information
about [rich messages](https://api.slack.com/docs/message-attachments) can be found in the
[Slack API documentation](https://api.slack.com/). The message format is exactly the same
as documented (at least currently!) so can be directly cut & pasted into Slack's
[message builder tool](https://api.slack.com/docs/messages/builder), e.g.:
[this is the below message, verbatim in the tool](https://api.slack.com/docs/messages/builder?msg=%7B%22channel%22%3A%22ABCDEFG%22%2C%22attachments%22%3A%5B%7B%22fallback%22%3A%22The%20deploy%20of%20%60YOURAPP.0.1.641%60%20to%20%60staging%60%20has%20completed%20with%20a%20status%20of%20*success*.%22%2C%22color%22%3A%22%2336a64f%22%2C%22title%22%3A%22Deployment%20Update%22%2C%22text%22%3A%22%3C%40USER%3E%2C%20the%20deploy%20of%20%60YOURAPP.0.1.641%60%20to%20%60staging%60%20has%20completed%20with%20a%20status%20of%20*success*.%22%2C%22fields%22%3A%5B%7B%22title%22%3A%22Artefact%22%2C%22value%22%3A%22YOURAPP.0.1.641%22%2C%22short%22%3Atrue%7D%2C%7B%22title%22%3A%22Environment%22%2C%22value%22%3A%22staging%22%2C%22short%22%3Atrue%7D%5D%2C%22actions%22%3A%5B%7B%22type%22%3A%22button%22%2C%22text%22%3A%22View%20Logs%22%2C%22url%22%3A%22http%3A%2F%2Flogs.example.com%2Fstaging%2Fdeploy%2FvFZhU-1388%22%2C%22style%22%3A%22primary%22%7D%5D%7D%5D%7D)

```json
{
  "name": "announce_deployments",
  "description": "Announces deployments",
  "steps": [
    {
      "id": "on_deployment_success",
      "event": {
        "packName": "MyContinuousDeploymentToolPack",
        "name": "DeploymentSucceeded"
      },
      "command": {
        "packName": "Slack",
        "name": "SendRichMessage",
        "input": {
          "channel": "ABCDEFG",
          "attachments": [
            {
              "fallback": "The deploy of `YOURAPP.0.1.641` to `staging` has completed with a status of *success*.",
              "color": "#36a64f",
              "title": "Deployment Update",
              "text": "<@USER>, the deploy of `YOURAPP.0.1.641` to `staging` has completed with a status of *success*.",
              "fields": [
                {
                  "title": "Artefact",
                  "value": "YOURAPP.0.1.641",
                  "short": true
                },
                {
                  "title": "Environment",
                  "value": "staging",
                  "short": true
                }
              ],
              "actions": [
                {
                  "type": "button",
                  "text": "View Logs",
                  "url": "http://logs.example.com/staging/deploy/vFZhU-1388",
                  "style": "primary"
                }
              ]
            }
          ]
        }
      }
    }
  ]
}
```
