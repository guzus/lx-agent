package notifier

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
}

func NewTelegram(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{botToken: botToken, chatID: chatID}
}

func (t *TelegramNotifier) Send(ctx context.Context, message string) error {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	// Try markdown first for richer formatting.
	if err := t.send(ctx, u, message, "Markdown"); err == nil {
		return nil
	}

	// Fallback to plain text when Telegram rejects markdown entities.
	return t.send(ctx, u, message, "")
}

func (t *TelegramNotifier) send(ctx context.Context, endpoint, message, parseMode string) error {
	params := url.Values{
		"chat_id": {t.chatID},
		"text":    {message},
	}
	if parseMode != "" {
		params.Set("parse_mode", parseMode)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram error: %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}
