package notification

import (
	"repo/internal/say"
	"repo/internal/util"

	"github.com/slack-go/slack"
)

const envSlackApiToken string = "REPOW_SLACK_API_TOKEN"
const envSlackChannelId string = "REPOW_SLACK_CHANNEL_ID"

// For api-token, go to https://api.slack.com and create a new applicaton and search for oauth api token.
// scopes: chat:write, chat:write.public
// channel must be public

func NotifyInvalidRepository(remotePath string, errorMessage string) {
	slackApiToken := util.GetEnv(envSlackApiToken, "")
	slackChannelId := util.GetEnv(envSlackChannelId, "")
	if slackApiToken != "" && slackChannelId != "" {
		api := slack.New(slackApiToken)
		message := ":large_blue_circle: repo.yaml for *" + remotePath + "* was invalid: " + errorMessage
		respChannel, respTimestamp, err := api.PostMessage(slackChannelId,
			slack.MsgOptionText(message, false),
			slack.MsgOptionAsUser(true))
		if err != nil {
			say.Error("Unable to send slack message: %v", err)
			return
		}
		say.Verbose("Send slack message @ %s to %s", respTimestamp, respChannel)
	}
}
