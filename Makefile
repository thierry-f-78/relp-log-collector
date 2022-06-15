all: relp-log-collector

relp-log-collector:
	go build -o relp-log-collector ./cmd/relp-log-collector

.PHONY: relp-log-collector
