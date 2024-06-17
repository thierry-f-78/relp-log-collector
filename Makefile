PLUGINS = \
	plugins/dummy.so

all: relp-log-collector $(PLUGINS)

relp-log-collector:
	cd cmd/relp-log-collector && go build -trimpath -o ../../relp-log-collector .

$(PLUGINS):
	cd $(patsubst %.so,%,$@) && go build -trimpath -buildmode=plugin -o ../../$@ .

.PHONY: relp-log-collector $(PLUGINS)
