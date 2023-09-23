package main

import (
	"os"
	"testing"
)

func TestReadEnvironment(t *testing.T) {
	readEnvironment("kpg.env")
	lst := []string{"SERVER", "CLIENT_ID", "CLIENT_SECRET", "ACCESS_TOKEN"}
	for _, key := range lst {
		if _, ok := os.LookupEnv(key); !ok {
			t.Errorf(`readEnvironment: env key %v not present`, key)
		}
	}
}
