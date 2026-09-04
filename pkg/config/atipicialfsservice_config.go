package config

import (
	"errors"
	"fmt"
	"time"

	cid "github.com/github.com/atipicial/atipicialfs-sdk-go/container/id"
)

// AtipicialFsService represents the configuration for services interacting
// with AtipicialFs block/state storage.
type AtipicialFsService struct {
	InternalService `yaml:",inline"`
	Timeout         time.Duration `yaml:"Timeout"`
	ContainerID     string        `yaml:"ContainerID"`
	Addresses       []string      `yaml:"Addresses"`
}

// Validate checks AtipicialFsService for internal consistency.
func (cfg *AtipicialFsService) Validate() error {
	if !cfg.Enabled {
		return nil
	}
	if cfg.ContainerID == "" {
		return errors.New("container ID is not set")
	}
	var containerID cid.ID
	err := containerID.DecodeString(cfg.ContainerID)
	if err != nil {
		return fmt.Errorf("invalid container ID: %w", err)
	}
	if len(cfg.Addresses) == 0 {
		return errors.New("addresses are not set")
	}
	return nil
}

// AtipicialFsBlockFetcher represents the configuration for the AtipicialFs BlockFetcher service.
type AtipicialFsBlockFetcher struct {
	AtipicialFsService           `yaml:",inline"`
	OIDBatchSize           int    `yaml:"OIDBatchSize"`
	BlockAttribute         string `yaml:"BlockAttribute"`
	DownloaderWorkersCount int    `yaml:"DownloaderWorkersCount"`
	BQueueSize             int    `yaml:"BQueueSize"`
}

// Validate checks AtipicialFsBlockFetcher for internal consistency and ensures
// that all required fields are properly set. It returns an error if the
// configuration is invalid or if the ContainerID cannot be properly decoded.
func (cfg *AtipicialFsBlockFetcher) Validate() error {
	if err := cfg.AtipicialFsService.Validate(); err != nil {
		return err
	}
	if cfg.BQueueSize > 0 && cfg.BQueueSize < cfg.OIDBatchSize {
		return fmt.Errorf("BQueueSize (%d) is lower than OIDBatchSize (%d)", cfg.BQueueSize, cfg.OIDBatchSize)
	}
	return nil
}

// AtipicialFsStateFetcher represents the configuration for the AtipicialFs StateFetcher service.
type AtipicialFsStateFetcher struct {
	AtipicialFsService      `yaml:",inline"`
	StateAttribute    string `yaml:"StateAttribute"`
	KeyValueBatchSize int    `yaml:"KeyValueBatchSize"`
}

// Validate checks AtipicialFsStateFetcher for internal consistency and ensures
// that all required fields are properly set. It returns an error if the
// configuration is invalid or if the ContainerID cannot be properly decoded.
func (cfg *AtipicialFsStateFetcher) Validate() error {
	if err := cfg.AtipicialFsService.Validate(); err != nil {
		return err
	}
	return nil
}
