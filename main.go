package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	_ "github.com/joho/godotenv/autoload"
	"google.golang.org/genai"
)

const template = `SYSTEM INSTRUCTIONS

ROLE:
You are an expert linguist and professional translator.

TASK:
Detect the language of the user's input and translate it to [%s].

GUIDELINES:
1. Accuracy: Preserve the original meaning and nuance. Do not add or remove information.
2. Tone: Maintain the [TONE: e.g., formal / friendly / authoritative] tone of the original text.
3. Idioms: Do not translate idioms literally. Replace them with an equivalent idiom or phrase in the target language that conveys the same meaning.
4. Cultural Nuance: Adapt cultural references so they make sense to a native speaker of the target language.
5. False Friends: Be vigilant about "false friends" (words that look similar but have different meanings) and ensure the correct term is used.
6. References: Make use of the references that has been mentioned and leave them where they are.

FORMATTING RULES:
- Output ONLY the translated text.
- Do not include "Here is the translation:" or quotes around the result.
- Retain original formatting (line breaks, bullet points) if present.

INPUT TEXT
[%s]`

type botHandler struct {
    bot    *gotgbot.Bot
    client *genai.Client
    logger *slog.Logger
}

func (h *botHandler) StartCommand(b *gotgbot.Bot, ctx *ext.Context) error {
    _, err := b.SendMessage(ctx.Message.From.Id, "Hello\\! I'm your friendly AI translator bot\\.\n\nTo get started, just send me a message like this:\n`Spanish Hello world`\n\nI automatically detect the language of your text\\! For more details, use the /help command\\.", &gotgbot.SendMessageOpts{
        ParseMode: "MarkdownV2",
    })
    if err != nil {
        return fmt.Errorf("failed to send start message: %w", err)
    }
    return nil
}

func (h *botHandler) HelpCommand(b *gotgbot.Bot, ctx *ext.Context) error {
    _, err := b.SendMessage(ctx.Message.From.Id, "To use me, send a message in the following format:\n\n`targetLanguage The text you want to translate`\n\n*I automatically detect the source language\\.*\n\nFor example:\n`French How are you?`\n\nYou can also use language codes:\n`es ¿Cómo estás?`", &gotgbot.SendMessageOpts{
        ParseMode: "MarkdownV2",
    })
    if err != nil {
        return fmt.Errorf("failed to send help message: %w", err)
    }
    return nil
}

func (h *botHandler) getTranslation(ctx context.Context, target, text string) (string, error) {
    res, err := h.client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(fmt.Sprintf(template, target, text)), nil)
    if err != nil {
        return "", err
    }
    return res.Text(), nil
}

func (h *botHandler) MessageHandler(b *gotgbot.Bot, ctx *ext.Context) error {
    parts := strings.SplitN(ctx.Message.Text, " ", 2)
    if len(parts) < 2 {
        _, err := b.SendMessage(ctx.Message.From.Id, "Invalid format\\. Please use the format: `targetLanguage text`\\.\n\nFor more information, use the /help command\\.", &gotgbot.SendMessageOpts{
            ParseMode: "MarkdownV2",
        })
        if err != nil {
            return fmt.Errorf("failed to send invalid format message: %w", err)
        }
        return nil
    }

    targetLanguage := parts[0]
    text := parts[1]

    _, err := h.bot.SendChatAction(ctx.Message.From.Id, "typing", &gotgbot.SendChatActionOpts{})
    if err != nil {
        return err
    }

    translated, err := h.getTranslation(context.Background(), targetLanguage, text)
    if err != nil {
        return err
    }

    _, err = b.SendMessage(ctx.Message.From.Id, translated, &gotgbot.SendMessageOpts{})
    if err != nil {
        return err
    }
    return nil
}

func main() {
    client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
        APIKey: os.Getenv("GEMINI_TOKEN"),
    })
    if err != nil {
        log.Fatalf("error creating gemini client")
    }
	bot, err := gotgbot.NewBot(os.Getenv("TELEGRAM_BOT_TOKEN"), &gotgbot.BotOpts{})
	if err != nil {
		log.Fatalf("error creating bot instance: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			logger.Error("an error occurred while handling update", "error", err)
			return ext.DispatcherActionNoop
		},
		MaxRoutines: ext.DefaultMaxRoutines,
	})

	updater := ext.NewUpdater(dispatcher, &ext.UpdaterOpts{})

	h := &botHandler{bot: bot, client: client, logger: logger}

	dispatcher.AddHandler(handlers.NewCommand("start", h.StartCommand))
	dispatcher.AddHandler(handlers.NewCommand("help", h.HelpCommand))
	dispatcher.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool { return !strings.HasPrefix(msg.Text, "/") }, h.MessageHandler))

	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		logger.Error("error getting update", "error", err)
	}
	logger.Info("Bot has started...", "bot_username", bot.Username)
	updater.Idle()
}