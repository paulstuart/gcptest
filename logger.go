// Copyright 2016 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

// Sample simplelog writes some entries, lists them, then deletes the log.
package main

import (
	"fmt"
	"log"

	"cloud.google.com/go/logging"
	"cloud.google.com/go/logging/logadmin"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

type gLogger struct {
	client *logging.Client
	admin  *logadmin.Client
}

func newClient(projID string) (*gLogger, error) {
	ctx := context.Background()
	client, err := logging.NewClient(ctx, projID)
	if err != nil {
		log.Fatalf("Failed to create logging client: %v", err)
		return nil, err
	}

	admin, err := logadmin.NewClient(ctx, projID)
	if err != nil {
		log.Fatalf("Failed to create logadmin client: %v", err)
		return nil, err
	}

	client.OnError = func(err error) {
		// Print an error to the local log.
		// For example, if Flush() failed.
		log.Printf("client.OnError: %v", err)
	}
	return &gLogger{client, admin}, nil
}

func (l *gLogger) writeEntry(logName, msg string) {
	logger := l.client.Logger(logName)

	infolog := logger.StandardLogger(logging.Info)
	infolog.Printf(msg)
	logger.Flush() // Ensure the entry is written.
}

func (l *gLogger) structuredWrite(logName string, payload interface{}) {
	logger := l.client.Logger(logName)

	logger.Log(logging.Entry{
		Payload:  payload,
		Severity: logging.Debug,
	})
	logger.Flush()
}

func (l *gLogger) deleteLog(logName string) error {
	ctx := context.Background()
	return l.admin.DeleteLog(ctx, logName)
}

func (l *gLogger) getEntries(projID, logName string) ([]*logging.Entry, error) {
	ctx := context.Background()

	var entries []*logging.Entry
	iter := l.admin.Entries(ctx,
		// Only get entries from the log-example log.
		logadmin.Filter(fmt.Sprintf(`logName = "projects/%s/logs/%s"`, projID, logName)),
		// Get most recent entries first.
		logadmin.NewestFirst(),
	)

	// Fetch the most recent 20 entries.
	for len(entries) < 20 {
		entry, err := iter.Next()
		if err == iterator.Done {
			return entries, nil
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
