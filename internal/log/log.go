package log

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/magicpool-co/pool/internal/telegram"
)

type Logger struct {
	MainLog   zerolog.Logger
	App       string
	Env       string
	Region    string
	LabelKeys []string
	ExitChan  chan os.Signal
	Telegram  *telegram.Client
}

func New(args map[string]string, application string, telegramClient *telegram.Client) (*Logger, error) {
	switch args["LOG_LEVEL"] {
	case "DEBUG":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "INFO":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "ERROR":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimestampFunc = func() time.Time {
		return time.Now().UTC()
	}

	output := os.Stdout

	var err error
	if logPath, ok := args["LOG_PATH"]; ok {
		output, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
	}

	logger := &Logger{
		MainLog:  zerolog.New(output).With().Timestamp().Logger(),
		App:      application,
		Env:      args["ENVIRONMENT"],
		Region:   args["AWS_REGION"],
		ExitChan: make(chan os.Signal),
		Telegram: telegramClient,
	}

	return logger, nil
}

func (l *Logger) label(event *zerolog.Event, labels []string) {
	if len(labels) == len(l.LabelKeys) {
		for i, label := range labels {
			event.Str(l.LabelKeys[i], label)
		}
	}
}

func (l *Logger) NotifyFarm(title, body string) {
	event := l.MainLog.Info().Str("app", "farm").Str("env", l.Env).Str("region", l.Region).Str("body", body)
	event.Msg(title)
}

func (l *Logger) Info(message string, labels ...string) {
	event := l.MainLog.Info().Str("app", l.App).Str("region", l.Region).Str("env", l.Env)
	l.label(event, labels)
	event.Msg(message)
}

func (l *Logger) Debug(message string, labels ...string) {
	event := l.MainLog.Debug().Str("app", l.App).Str("region", l.Region).Str("env", l.Env)
	l.label(event, labels)
	event.Msg(message)
}

func (l *Logger) Error(err error, labels ...string) {
	event := l.MainLog.Error().Stack().Err(err).Str("app", l.App).Str("env", l.Env).Str("region", l.Region)
	l.label(event, labels)
	event.Msg("")
}

func (l *Logger) Fatal(err error, labels ...string) {
	if l.Telegram != nil {
		l.Telegram.SendFatal(err.Error(), l.App, l.Env)
	}

	event := l.MainLog.WithLevel(zerolog.FatalLevel).Stack().Err(err).Str("app", l.App).Str("env", l.Env).Str("region", l.Region)
	l.label(event, labels)
	event.Msg("")
}

func (l *Logger) Panic(r interface{}, trace string, labels ...string) {
	var err error

	switch x := r.(type) {
	case string:
		err = fmt.Errorf(x)
	case error:
		err = x
	default:
		err = fmt.Errorf("unknown panic")
	}

	if l.Telegram != nil {
		err := l.Telegram.SendPanic(err.Error(), l.App, l.Env)
		if err != nil {
			l.Error(err)
		}
	}

	event := l.MainLog.WithLevel(zerolog.PanicLevel).Stack().Err(err).Str("trace", trace).Str("app", l.App).Str("env", l.Env).Str("region", l.Region)
	l.label(event, labels)
	event.Msg("")

	l.ExitChan <- syscall.SIGTERM
}
