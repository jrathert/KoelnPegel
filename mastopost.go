package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mattn/go-mastodon"
)

func postToMastodon(statusText string) (mastodon.ID, error) {
	if false {
		c := mastodon.NewClient(&mastodon.Config{
			Server:       os.Getenv("SERVER"),
			ClientID:     os.Getenv("CLIENT_ID"),
			ClientSecret: os.Getenv("CLIENT_SECRET"),
			AccessToken:  os.Getenv("ACCESS_TOKEN"),
		})
		toot := &mastodon.Toot{
			Status: statusText,
		}
		status, err := c.PostStatus(context.Background(), toot)
		if err != nil {
			var null mastodon.ID
			return null, err
		}
		return status.ID, nil
	} else {
		fmt.Printf("Mastodon post:\n%v\n", statusText)
		return mastodon.ID("42"), nil
	}
}
