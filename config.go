package logtail

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
)

const DefaultServerPort = 54321

var (
	file          = flag.String("file", "", "config file")
	port          = flag.Int("port", DefaultServerPort, "tail port")
	command       = flag.String("cmd", "", "tail command")
	matchContains = flag.String("match_contains", "", "a containing string")
	dingUrl       = flag.String("ding_alert_url", "", "dingding alert url")
	webhookUrl    = flag.String("webhook_alert_url", "", "webhook alert url")
)

type Config struct {
	Port    int             `json:"port"`
	Servers []*ServerConfig `json:"serverDB"`
}

type ServerConfig struct {
	ID      string          `json:"id"`
	Command string          `json:"command"`
	Routers []*RouterConfig `json:"routers"`
}

type RouterConfig struct {
	Filters   []*FilterConfig   `json:"filters"`
	Transfers []*TransferConfig `json:"transfers"`
}

type FilterConfig struct {
	MatchContains string `json:"match_contains"`
}

type TransferConfig struct {
	DingURL    string `json:"ding_url"`
	WebhookURL string `json:"webhook_url"`
}

func parseConfig() (*Config, error) {
	config, err := readConfig()
	if err != nil {
		return nil, err
	}
	if config.Port == 0 {
		config.Port = DefaultServerPort
	}
	if len(config.Servers) == 0 {
		return nil, errors.New("no server config")
	}
	for _, server := range config.Servers {
		if err := validateServerConfig(server); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func readConfig() (*Config, error) {
	config := &Config{}
	if *file != "" {
		data, err := ioutil.ReadFile(*file)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}
		return config, nil
	}

	config.Port = *port
	serverConfig := &ServerConfig{
		ID: DefaultServerId,
	}

	config.Servers = append(config.Servers, serverConfig)
	serverConfig.Command = *command

	if *dingUrl == "" && *webhookUrl == "" && *matchContains == "" {
		return config, nil
	}

	routerConfig := &RouterConfig{}
	serverConfig.Routers = append(serverConfig.Routers, routerConfig)

	if *matchContains != "" {
		routerConfig.Filters = append(routerConfig.Filters, &FilterConfig{
			MatchContains: *matchContains,
		})
	}
	if *dingUrl != "" {
		routerConfig.Transfers = append(routerConfig.Transfers, &TransferConfig{
			DingURL: *dingUrl,
		})
	}
	if *webhookUrl != "" {
		routerConfig.Transfers = append(routerConfig.Transfers, &TransferConfig{
			WebhookURL: *webhookUrl,
		})
	}

	return config, nil
}

func validateServerConfig(server *ServerConfig) error {
	if server.ID == "" {
		return errors.New("server id is nil")
	}
	if server.Command == "" {
		return errors.New("server command is nil")
	}
	if len(server.Routers) > 0 {
		for _, router := range server.Routers {
			if err := validateRouterConfig(router); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRouterConfig(router *RouterConfig) error {
	if len(router.Filters) > 0 {
		for _, filter := range router.Filters {
			if err := validateFilterConfig(filter); err != nil {
				return err
			}
		}
	}
	if len(router.Transfers) > 0 {
		for _, transfer := range router.Transfers {
			if err := validateTransferConfig(transfer); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateTransferConfig(transfer *TransferConfig) error {
	if transfer.DingURL == "" && transfer.WebhookURL == "" {
		return errors.New("transfer url config is nil")
	}
	return nil
}

func validateFilterConfig(filter *FilterConfig) error {
	if filter.MatchContains == "" {
		return errors.New("match contains is nil")
	}
	return nil
}
