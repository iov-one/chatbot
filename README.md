# Chatbot

Chatbot is a Slack bot to deploy things to Kubernetes

## Requirements

This assumes you already have a kubernetes cluster. Tested with kubectl v1.11.1.

## Setup

To run Chatbot on Slack, first you need to [create a new bot](https://my.slack.com/services/new/bot) user integration on Slack and get the `token` (See [Slack bot users](https://api.slack.com/bot-users) for more details).

Then you need to know the channel ids where you want to run the Chatbot. You can get them on `https://slack.com/api/channels.list?token={REPLACE WITH YOUR TOKEN}`

## How to run

### Binary
You just need to supply one environment variable, which is the slack token you've created earlier:  
`CHATBOT_SLACK_TOKEN=your_token go run cmd/bot/main.go`


## How to use
After creating and setting up a bot - invite it to your desired Slack channel(s) by mentioning it.

### Supported commands
* `!deploy your_app your_container your/docker:image`, currently supports deployments and statefulsets, where `your_app`  
is a statefulset/deployment name, `your_container` is a container name within a pod and image is the image you want  
to deploy. Note, that it automatically forces redeploy even if the image name is the same as the last deployed one  
but if you want to force pull - you need to specify `imagePullPolicy: Always` for your container in the manifest.
* A link with concrete commands per artifact can be found [here](https://github.com/iov-one/devnet-operations/blob/master/README.md#deploying-current-artifacts-with-chatbot)
