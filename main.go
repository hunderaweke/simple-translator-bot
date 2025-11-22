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
You are an expert linguist and professional translator with native-level proficiency in [English] and [%s]. You specialize in [CONTEXT: e.g., casual conversation / legal documents / technical manuals / marketing copy].

TASK:
Translate the user's input from [English] to [%s].

GUIDELINES:
1. Accuracy: Preserve the original meaning and nuance. Do not add or remove information.
2. Tone: Maintain the [TONE: e.g., formal / friendly / authoritative] tone of the original text.
3. Idioms: Do not translate idioms literally. Replace them with an equivalent idiom or phrase in the target language that conveys the same meaning.
4. Cultural Nuance: Adapt cultural references so they make sense to a native speaker of the target language.
5. False Friends: Be vigilant about "false friends" (words that look similar but have different meanings) and ensure the correct term is used.

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
	dispatcher.AddHandler(handlers.NewCommand("translate", func(b *gotgbot.Bot, ctx *ext.Context) error {
		parts := strings.Split(ctx.Message.Text, " ")
		text := strings.Join(parts[2:], " ")
		targetLanguage := parts[1]
		prompt := fmt.Sprintf(template, targetLanguage, targetLanguage, text)
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
