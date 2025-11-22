# 🌍 Translator Bot ✨

A powerful, AI-driven Telegram bot that provides professional-grade translations right within your chat. Break down language barriers and communicate effortlessly with anyone, anywhere!

[![Go](https://img.shields.io/badge/Go-1.18+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Telegram](https://img.shields.io/badge/Telegram-Bot-26A5E4?style=for-the-badge&logo=telegram)](https://telegram.org/)
[![Gemini AI](https://img.shields.io/badge/Gemini_AI-Google-4285F4?style=for-the-badge&logo=google)](https://ai.google.dev/)

## 🚀 Features

- **High-Quality Translations**: Leverages Google's state-of-the-art Gemini AI for accurate and nuanced translations.
- **Multi-Language Support**: Translate between a vast number of languages.
- **Context-Aware**: Provides translations that understand context, tone, and idiomatic expressions.
- **Easy to Use**: A simple and intuitive command structure.
- **Fast and Responsive**: Get your translations in seconds.

## 📖 Usage

Interacting with the bot is simple. Just send a message in the following format:

```
sourceLanguage->targetLanguage The text you want to translate
```

**For example:**

To translate "Hello, how are you?" from English to French, you would send:

```
English->French Hello, how are you?
```

The bot will then reply with the translated text:

```
Bonjour, comment ça va ?
```

It's that easy! 🎉

## 🛠️ Setup & Installation

To run your own instance of the Translator Bot, follow these steps.

### Prerequisites

- [Go](https://go.dev/doc/install) (version 1.18 or higher)
- A [Telegram Bot Token](https://core.telegram.org/bots#6-botfather)
- A [Gemini API Key](https://makersuite.google.com/app/apikey)

### Installation

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/your-username/translator-bot.git
    cd translator-bot
    ```

2.  **Set up your environment variables:**
    Create a file named `.env` in the root of the project directory and add your credentials:

    ```
    TELEGRAM_BOT_TOKEN="YOUR_TELEGRAM_BOT_TOKEN"
    GEMINI_TOKEN="YOUR_GEMINI_API_KEY"
    ```

    The bot uses `godotenv` to automatically load these variables.

3.  **Install dependencies:**

    ```bash
    go mod tidy
    ```

4.  **Run the bot:**
    ```bash
    go run main.go
    ```

Your bot is now live and ready to start translating! 🤖

## ☁️ Deployment

You can deploy this bot to any cloud service that supports Go applications. Here are a few popular options:

- **Heroku**: A simple platform-as-a-service (PaaS) that makes deployment easy.
- **DigitalOcean / AWS / Google Cloud**: You can run the bot on a Virtual Private Server (VPS) for more control.
- **Serverless Functions**: For a more cost-effective and scalable solution, you can adapt the code to run on services like AWS Lambda or Google Cloud Functions.

When deploying, make sure to set the environment variables (`TELEGRAM_BOT_TOKEN` and `GEMINI_TOKEN`) in your hosting provider's configuration settings.

## 🤝 Contributing

Contributions are welcome! If you have ideas for new features, improvements, or bug fixes, feel free to:

1.  Fork the repository.
2.  Create a new branch (`git checkout -b feature/YourFeature`).
3.  Make your changes.
4.  Commit your changes (`git commit -m 'Add some feature'`).
5.  Push to the branch (`git push origin feature/YourFeature`).
6.  Open a Pull Request.

## 📜 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
