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
	ErrNoServerConfig   = errors.New("no server config")
	ErrServerIDNil      = errors.New("server id is nil")
	ErrServerCommandNil = errors.New("server command is nil")
	ErrTransURLNil      = errors.New("transfer url is nil")
	ErrTransTypeNil     = errors.New("transfer type is nil")
	ErrTransTypeInvalid = errors.New("invalid transfer type")
)

type Config struct {
	Port           int             `json:"port"`
	DefaultFormat  *Format         `json:"default_format"`
	Servers        []*ServerConfig `json:"servers"`
	DefaultRouters []*RouterConfig `json:"default_routers"`
	GlobalRouters  []*RouterConfig `json:"global_routers"`
}

type ServerConfig struct {
	ID         string          `json:"id"`
	Command    string          `json:"command"`     // single command
	Commands   string          `json:"commands"`    // multiple commands split by new line
	CommandGen string          `json:"command_gen"` // command to generate multiple commands split by new line
	Format     *Format         `json:"format"`
	Routers    []*RouterConfig `json:"routers"`
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

func parseConfig() (cfg *Config, parseErr error) {
	defer func() {
		if err := recover(); err != nil {
			parseErr = err.(error)
		}
	}()

	var (
		file          = flag.String("file", "", "config file")
		port          = flag.Int("port", DefaultServerPort, "tail port")
		command       = flag.String("cmd", "", "tail command")
		matchContains = flag.String("match-contains", "", "a containing string")
		dingURL       = flag.String("ding-url", "", "dingding url")
		webhookURL    = flag.String("webhook-url", "", "webhook url")
	)

	flag.Parse()

	config, err := readConfig(*file, *port, *command, *matchContains, *dingURL, *webhookURL)

	if err != nil {
		return nil, err
	}

	if config.Port == 0 {
		config.Port = DefaultServerPort
	}

	if len(config.Servers) == 0 {
		return nil, ErrNoServerConfig
	}

	for _, server := range config.Servers {
		if err := validateServerConfig(server); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func readConfig(file string, port int, command, matchContains, dingURL, webhookURL string) (*Config, error) {
	config := &Config{}

	if file != "" {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			return nil, err
		}

		if err := json.Unmarshal(data, config); err != nil {
			return nil, err
		}

		return config, nil
	}

	config.Port = port
	serverConfig := &ServerConfig{
		ID: DefaultServerID,
	}

	config.Servers = append(config.Servers, serverConfig)
	serverConfig.Command = command

	if dingURL == "" && webhookURL == "" && matchContains == "" {
		return config, nil
	}

	routerConfig := &RouterConfig{}
	serverConfig.Routers = append(serverConfig.Routers, routerConfig)

	if matchContains != "" {
		routerConfig.Matchers = append(routerConfig.Matchers, &MatcherConfig{
			MatchContains: matchContains,
		})
	}

	if dingURL != "" {
		routerConfig.Transfers = append(routerConfig.Transfers, &TransferConfig{
			Type: TransferTypeDing,
			URL:  dingURL,
		})
	}

	if webhookURL != "" {
		routerConfig.Transfers = append(routerConfig.Transfers, &TransferConfig{
			Type: TransferTypeWebhook,
			URL:  webhookURL,
		})
	}

	return config, nil
}

func validateServerConfig(server *ServerConfig) error {
	if server.ID == "" {
		return ErrServerIDNil
	}

	if server.Command == "" && server.Commands == "" && server.CommandGen == "" {
		return ErrServerCommandNil
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
		return ErrTransTypeNil
	}

	if transfer.Type != TransferTypeWebhook && transfer.Type != TransferTypeDing && transfer.Type != TransferTypeConsole {
		return fmt.Errorf("%w: %s", ErrTransTypeInvalid, transfer.Type)
	}

	if transfer.Type == TransferTypeWebhook || transfer.Type == TransferTypeDing {
		if transfer.URL == "" {
			return ErrTransURLNil
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
