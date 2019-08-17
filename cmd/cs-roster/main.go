package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	internalhttp "github.com/payfazz/go-skeleton/internal/http"
	"github.com/payfazz/go-skeleton/internal/notif"
)

func getEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return defaultValue
}

var (
	dev = flag.Bool(
		"dev", false, "enable the dev mode, you can bypass the auth process in dev mode.",
	)
	port = flag.String(
		"port", getEnv("PAYFAZZ_GO_SKELETON_PORT", "8080"), "http port to listen on",
	)
	slackToken = flag.String(
		"slacktoken", getEnv("PAYFAZZ_GO_SKELETON_SLACK_TOKEN", ""), "slack api token",
	)
	slackChannel = flag.String(
		"slackchannel", getEnv("PAYFAZZ_GO_SKELETON_SLACK_CHANNEL", "#alert-coupon"), "slack alert channel",
	)
)

func main() {
	flag.Parse()

	s := &http.Server{
		Addr: ":" + *port,
		Handler: internalhttp.NewServer(
			*dev,
			internalhttp.NewResponder(notif.NewSlackNotifier(notif.SlackNotifierConfig{
				Token:   *slackToken,
				Channel: *slackChannel,
			})),
		),
	}

	go func() {
		if *dev {
			log.Printf("Listening on :%v ... [Dev Mode]", *port)
		} else {
			log.Printf("Listening on :%v ...", *port)
		}
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error listening on :%v: %v", *port, err)
		}
	}()

	// graceful shutdown
	//

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	<-quit
	log.Println("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}
