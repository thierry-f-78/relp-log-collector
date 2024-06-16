package main

import "fmt"

import "github.com/thierry-f-78/relp-log-collector/pkg/backend"

func Init() error {
	fmt.Printf("dummy plugin successufully initialized\n")
	return nil
}

func NewBatch() error {
	return nil
}

func Pick(m *backend.Message) (bool, error) {
	return false, nil
}

func Flush() error {
	return nil
}
