package cache

import (
	"errors"
	"github.com/ExpediaGroup/flyte-slack/types"
	"github.com/HotelsDotCom/go-logger"
	"time"
)

var (
	errNoInit        = errors.New("cache not initialized")
	errNoSuchChannel = errors.New("can't find channel with such name")
)

type Config struct {
	RenewConversationListFrequency time.Duration
}

// slackClient exposes only methods needed for cache
type slackClient interface {
	GetConversations() ([]types.Conversation, error)
}

type Cache interface {
	GetChannelID(channelName string, client slackClient) (*types.Conversation, error)
}

type cache struct {
	cfg *Config
	// conversationsList maps channel names to other channel data
	conversationsList       map[string]types.Conversation
	conversationListUpdated *time.Time
}

func (c *cache) isConversationListUpdateNeeded() bool {
	return c.conversationListUpdated == nil ||
		time.Since(*c.conversationListUpdated) > c.cfg.RenewConversationListFrequency
}

func (c *cache) updateConversationList(client slackClient) error {
	conv, err := client.GetConversations()
	if err != nil {
		return err
	}

	n := time.Now()
	c.conversationListUpdated = &n

	// clear cache
	for k := range c.conversationsList {
		delete(c.conversationsList, k)
	}

	// add new values
	for i := range conv {
		c.conversationsList[conv[i].Name] = conv[i]
	}

	return nil
}

// GetChannelID will get channel ID from cache or make relevant API call if
// cache is empty or time to renew cache has come (defined by config)
func (c *cache) GetChannelID(channelName string, client slackClient) (*types.Conversation, error) {
	if c.isConversationListUpdateNeeded() {
		err := c.updateConversationList(client)
		if err != nil {
			logger.Errorf("can't update conversation list cache: %s", err)
		}
	}

	if out, ok := c.conversationsList[channelName]; !ok {
		return nil, errNoSuchChannel
	} else {
		return &out, nil
	}
}

func New(config *Config) Cache {
	return &cache{
		cfg:               config,
		conversationsList: make(map[string]types.Conversation),
	}
}
