package main

import (
	"os"

	"github.com/l0wl3vel/bunnystorage-go/cmd/bunnystoragectl/internal/app"
)

func main() {
	os.Exit(app.Run())
}
