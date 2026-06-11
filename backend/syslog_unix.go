//go:build !windows

package main

import (
	"io"
	"log"
	"log/syslog"
	"os"
)

func setupSyslog(tag string) {
	w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, tag)
	if err != nil {
		log.Printf("syslog unavailable: %v", err)
		return
	}
	log.SetOutput(io.MultiWriter(os.Stderr, w))
}
