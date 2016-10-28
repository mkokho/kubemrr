package app

import log "github.com/Sirupsen/logrus"

func init() {
	log.SetLevel(log.InfoLevel)
	log.SetFormatter(&log.JSONFormatter{TimestampFormat: "2006-01-02T15:04:05.000"})
}

func enableDebug() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{})
}
