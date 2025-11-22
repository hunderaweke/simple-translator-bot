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
You are an expert linguist and professional translator with native-level proficiency in [%s] and [%s]. You specialize in [CONTEXT: e.g., casual conversation / legal documents / technical manuals / marketing copy].

TASK:
Translate the user's input from [%s] to [%s].

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
	dispatcher.AddHandler(handlers.NewCommand("start", func(b *gotgbot.Bot, ctx *ext.Context) error {
		_, err := b.SendMessage(ctx.Message.From.Id, "Hello\\! I'm your friendly AI translator bot\\.\n\nTo get started, just send me a message like this:\n`English->Spanish Hello world`\n\nFor more details, use the /help command\\.", &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})
		if err != nil {
			return fmt.Errorf("failed to send start message: %w", err)
		}
		return nil
	}))
	dispatcher.AddHandler(handlers.NewCommand("help", func(b *gotgbot.Bot, ctx *ext.Context) error {
		_, err := b.SendMessage(ctx.Message.From.Id, "To use me, send a message in the following format:\n\n`sourceLanguage->targetLanguage The text you want to translate`\n\nFor example:\n`English->French Hello, how are you?`", &gotgbot.SendMessageOpts{
			ParseMode: "MarkdownV2",
		})
		if err != nil {
			return fmt.Errorf("failed to send help message: %w", err)
		}
		return nil
	}))
	dispatcher.AddHandler(handlers.NewMessage(func(msg *gotgbot.Message) bool {
		return !strings.HasPrefix(msg.Text, "/")
	}, func(b *gotgbot.Bot, ctx *ext.Context) error {
		var text string
		var languages []string

		// New logic to handle spaces around "->"
		if strings.Contains(ctx.Message.Text, "->") {
			parts := strings.SplitN(ctx.Message.Text, "->", 2)
			langParts := strings.Fields(parts[0])
			if len(langParts) > 0 {
				sourceLang := langParts[len(langParts)-1]

				textParts := strings.Fields(parts[1])
				if len(textParts) > 0 {
					targetLang := textParts[0]
					languages = []string{sourceLang, targetLang}
					text = strings.Join(textParts[1:], " ")
				}
			}
		}

		if len(languages) != 2 || text == "" {
			_, err := b.SendMessage(ctx.Message.From.Id, "Invalid format. Please use the format: `sourceLanguage->targetLanguage text`.\n\nFor more information, use the /help command.", &gotgbot.SendMessageOpts{
				ParseMode: "MarkdownV2",
			})
			if err != nil {
				return fmt.Errorf("failed to send invalid format message: %w", err)
			}
			return nil
		}
		prompt := fmt.Sprintf(template, languages[0], languages[1], languages[0], languages[1], text)
		_, err := bot.SendChatAction(ctx.Message.From.Id, "typing", &gotgbot.SendChatActionOpts{})
		if err != nil {
			return err
		}
		done := make(chan interface{})
		defer close(done)
		type result struct {
			Error  error
			Result string
		}
		sendRequest := func(done <-chan interface{}, prompt string) chan result {
			resultChan := make(chan result)
			go func() {
				defer close(resultChan)
				var r result
				res, err := client.Models.GenerateContent(context.Background(), "gemini-2.5-flash", genai.Text(prompt), nil)
				if err != nil {
					r.Error = err
				}
				r.Result = res.Text()
				select {
				case <-done:
					return
				case resultChan <- r:
				}
			}()
			return resultChan
		}
		r := sendRequest(done, prompt)
		for res := range r {
			if res.Error != nil {
				return err
			}
			_, err = bot.SendMessage(ctx.Message.From.Id, res.Result, &gotgbot.SendMessageOpts{})
			if err != nil {
				return err
			}
		}
		return nil
	}))
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
		logger.Error("error getting update: ", err)
	}
	logger.Info("Bot has started...", "bot_username", bot.Username)
	updater.Idle()
}
