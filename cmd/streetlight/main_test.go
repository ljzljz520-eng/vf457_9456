package main

import (
	"streetlight/operations"
	"streetlight/storage"
	"streetlight/workflow"
	"testing"
)

func TestDemoEntryPoint(t *testing.T) {
	db, path, err := storage.OpenEphemeral()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	_ = path
	s := operations.NewService(db)
	if err = workflow.Seed(s); err != nil {
		t.Fatal(err)
	}
	if err = runDemo(s); err != nil {
		t.Fatal(err)
	}
}
