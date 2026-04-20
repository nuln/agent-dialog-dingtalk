// Package dingtalk integrates DingTalk (钉钉) as a dialog platform via Stream Mode.
package dingtalk

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	agent "github.com/nuln/agent-core"
	dingchatbot "github.com/open-dingtalk/dingtalk-stream-sdk-go/chatbot"
	dingclient "github.com/open-dingtalk/dingtalk-stream-sdk-go/client"
)

func init() {
	agent.RegisterPluginConfigSpec(agent.PluginConfigSpec{
		PluginName:  "dingtalk",
		PluginType:  "dialog",
		Description: "DingTalk robot dialog platform integration (Stream Mode)",
		AuthType:    agent.AuthTypeToken,
		Fields: []agent.ConfigField{
			{Key: "client_id", EnvVar: "DINGTALK_CLIENT_ID", Description: "DingTalk App Client ID", Required: true, Type: agent.ConfigFieldString},
			{Key: "client_secret", EnvVar: "DINGTALK_CLIENT_SECRET", Description: "DingTalk App Client Secret", Required: true, Type: agent.ConfigFieldSecret},
		},
	})

	agent.RegisterDialog("dingtalk", func(opts map[string]any) (agent.Dialog, error) {
		return New(opts)
	})
}

type replyContext struct {
	sessionWebhook string
	userID         string
}

// DingTalkDialog implements agent.Dialog for DingTalk chatbot via Stream Mode.
type DingTalkDialog struct {
	mu           sync.RWMutex
	clientID     string
	clientSecret string
	handler      agent.MessageHandler
	status       agent.DialogInstanceStatus
	cancel       context.CancelFunc
}

// New creates a DingTalkDialog from options.
func New(opts map[string]any) (*DingTalkDialog, error) {
	clientID, _ := opts["client_id"].(string)
	if clientID == "" {
		clientID = os.Getenv("DINGTALK_CLIENT_ID")
	}
	clientSecret, _ := opts["client_secret"].(string)
	if clientSecret == "" {
		clientSecret = os.Getenv("DINGTALK_CLIENT_SECRET")
	}
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("dingtalk: DINGTALK_CLIENT_ID and DINGTALK_CLIENT_SECRET are required")
	}
	return &DingTalkDialog{clientID: clientID, clientSecret: clientSecret}, nil
}

func (d *DingTalkDialog) Name() string { return "dingtalk" }

func (d *DingTalkDialog) Start(handler agent.MessageHandler) error {
	d.handler = handler

	ctx, cancel := context.WithCancel(context.Background())
	d.cancel = cancel

	cli := dingclient.NewStreamClient(dingclient.WithAppCredential(
		dingclient.NewAppCredentialConfig(d.clientID, d.clientSecret),
	))

	cli.RegisterChatBotCallbackRouter(func(c context.Context, df *dingchatbot.BotCallbackDataModel) ([]byte, error) {
		userID := df.SenderStaffId
		if userID == "" {
			userID = df.SenderCorpId
		}
		rctx := replyContext{sessionWebhook: df.SessionWebhook, userID: userID}
		sessionKey := fmt.Sprintf("dingtalk:%s", userID)
		msg := &agent.Message{
			SessionKey: sessionKey,
			UserID:     userID,
			Content:    df.Text.Content,
			ReplyCtx:   rctx,
		}
		d.mu.Lock()
		d.status.InboundAt = time.Now().UnixMilli()
		d.mu.Unlock()
		if d.handler != nil {
			d.handler(d, msg)
		}
		return nil, nil
	})

	go func() {
		if err := cli.Start(ctx); err != nil && ctx.Err() == nil {
			slog.Error("dingtalk stream client error", "err", err)
		}
	}()

	d.mu.Lock()
	d.status = agent.DialogInstanceStatus{
		ID:          "dingtalk",
		Status:      "connected",
		InboundAt:   time.Now().UnixMilli(),
		Description: "DingTalk stream mode connected",
	}
	d.mu.Unlock()
	slog.Info("dingtalk: dialog started")
	return nil
}

func (d *DingTalkDialog) Reply(ctx context.Context, replyCtx any, content string) error {
	return d.Send(ctx, replyCtx, content)
}

func (d *DingTalkDialog) Send(ctx context.Context, replyCtx any, content string) error {
	rctx, ok := replyCtx.(replyContext)
	if !ok {
		return fmt.Errorf("dingtalk: invalid reply context")
	}
	replier := dingchatbot.NewChatbotReplier()
	return replier.SimpleReplyText(ctx, rctx.sessionWebhook, []byte(content))
}

func (d *DingTalkDialog) Stop() error {
	if d.cancel != nil {
		d.cancel()
	}
	return nil
}

func (d *DingTalkDialog) Reload(opts map[string]any) error {
	_ = d.Stop()
	nd, err := New(opts)
	if err != nil {
		return err
	}
	d.clientID = nd.clientID
	d.clientSecret = nd.clientSecret
	return d.Start(d.handler)
}
