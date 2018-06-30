package main // import "github.com/mojlighetsministeriet/storage"

import (
	"os"

	"github.com/mojlighetsministeriet/storage/remote"
	"github.com/mojlighetsministeriet/utils"
)

func main() {
	useTLS := true
	if os.Getenv("TLS") == "disable" {
		useTLS = false
	}
	bodyLimit := utils.GetEnv("BODY_LIMIT", "5M")
	port := ":" + utils.GetEnv("PORT", "443")

	service := remote.NewService(useTLS, true, bodyLimit)
	service.Listen(port)
}
