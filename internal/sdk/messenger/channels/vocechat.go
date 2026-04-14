package channels

import "github.com/engigu/baihu-panel/internal/sdk/message"

type VoceChatChannel struct{ *BaseChannel }

func NewVoceChatChannel() Channel {
	return &VoceChatChannel{NewBaseChannel(ChannelVoceChat, []string{FormatTypeMarkdown})}
}

func (c *VoceChatChannel) Send(config ChannelConfig, msg *Message) (*Result, error) {
	server := config.GetString("server")
	apiKey := config.GetString("api_key")
	targetID := config.GetString("target_id")

	if server == "" || apiKey == "" || targetID == "" {
		return SendError("vocechat config missing: server, api_key and target_id are required"), nil
	}

	cli := message.VoceChat{
		Server:     server,
		APIKey:     apiKey,
		TargetType: config.GetString("target_type"), // defaults to "user" in SDK if not matched differently
		TargetID:   targetID,
	}

	res, err := cli.Request(msg.Title, msg.Text)
	if err != nil {
		return ErrorResult(string(res), err), nil
	}
	return SuccessResult(string(res)), nil
}
