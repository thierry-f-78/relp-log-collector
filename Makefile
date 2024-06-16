PLUGINS = \
	plugins/dummy.so

all: relp-log-collector $(PLUGINS)

relp-log-collector:
	go build -trimpath -o relp-log-collector ./cmd/relp-log-collector

$(PLUGINS):
	go build -trimpath -buildmode=plugin -o $@ ./$(patsubst %.so,%,$@)

.PHONY: relp-log-collector $(PLUGINS)
