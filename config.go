package logtail

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/vogo/logger"
)

const DefaultServerPort = 54321

var (
	file          = flag.String("file", "", "config file")
	port          = flag.Int("port", DefaultServerPort, "tail port")
	command       = flag.String("cmd", "", "tail command")
	matchContains = flag.String("match-contains", "", "a containing string")
	dingUrl       = flag.String("ding-url", "", "dingding url")
	webhookUrl    = flag.String("webhook-url", "", "webhook url")
)

type Config struct {
	Port           int             `json:"port"`
	Servers        []*ServerConfig `json:"servers"`
	DefaultRouters []*RouterConfig `json:"default_routers"`
	GlobalRouters  []*RouterConfig `json:"global_routers"`
}

type ServerConfig struct {
	ID      string          `json:"id"`
	Command string          `json:"command"`
	Routers []*RouterConfig `json:"routers"`
}

type RouterConfig struct {
	ID        string            `json:"id"`
	Matchers  []*MatcherConfig  `json:"matchers"`
	Transfers []*TransferConfig `json:"transfers"`
}

type MatcherConfig struct {
	MatchContains string `json:"match_contains"`
}

type TransferConfig struct {
	Type string `json:"type"`
	URL  string `json:"url"`
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
		routerConfig.Matchers = append(routerConfig.Matchers, &MatcherConfig{
			MatchContains: *matchContains,
		})
	}
	if *dingUrl != "" {
		routerConfig.Transfers = append(routerConfig.Transfers, &TransferConfig{
			Type: TransferTypeDing,
			URL:  *dingUrl,
		})
	}
	if *webhookUrl != "" {
		routerConfig.Transfers = append(routerConfig.Transfers, &TransferConfig{
			Type: TransferTypeWebhook,
			URL:  *webhookUrl,
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
	if err := validateMatchers(router.Matchers); err != nil {
		return err
	}
	if err := validateTransfers(router.Transfers); err != nil {
		return err
	}
	return nil
}

func validateMatchers(matchers []*MatcherConfig) error {
	if len(matchers) > 0 {
		for _, filter := range matchers {
			if err := validateMatchConfig(filter); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateTransfers(transfers []*TransferConfig) error {
	if len(transfers) > 0 {
		for _, transfer := range transfers {
			if err := validateTransferConfig(transfer); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateTransferConfig(transfer *TransferConfig) error {
	if transfer.Type == "" {
		return errors.New("transfer type is nil")
	}

	if transfer.Type != TransferTypeWebhook && transfer.Type != TransferTypeDing && transfer.Type != TransferTypeConsole {
		return fmt.Errorf("transfer type %s is invalid", transfer.Type)
	}

	if transfer.Type == TransferTypeWebhook || transfer.Type == TransferTypeDing {
		if transfer.URL == "" {
			return errors.New("transfer url is nil")
		}
	}

	return nil
}

func validateMatchConfig(config *MatcherConfig) error {
	if config.MatchContains == "" {
		logger.Debugf("match contains is nil")
	}
	return nil
}
