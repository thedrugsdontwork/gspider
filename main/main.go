package main

import (
	"GSpider"
	"go.uber.org/zap"
	"time"
)

func test() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	var sugar = logger.Sugar()
	sugar.Infow("failed to fetch URL",
		// Structured context as loosely typed key-value pairs.
		"url", "asdasdasdasdad",
		"attempt", 3,
		"backoff", time.Second,
	)
	sugar.Infof("Failed to fetch URL: %s", "asdadasdasdassss")
	logger.Info("failed to fetch URL",
		// Structured context as strongly typed Field values.
		zap.String("url", "asdasdaasda"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}
func main() {
	GSpider.TestForCrawler()
}
