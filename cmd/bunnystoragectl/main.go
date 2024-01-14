package main

import (
	"os"

	"jamesponddotco/cmd/bunnystoragectl/internal/app"
)

func main() {
	os.Exit(app.Run())
}
