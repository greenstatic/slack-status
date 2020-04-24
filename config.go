package main

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

type Config struct {
	Workspaces []Workspace
}

type Workspace struct {
	Name        string
	AccessToken string `yaml:"accessToken"`
	Groups      []Group
}

type Group string

func (c *Config) Parse(filepath string) error {
	contents, err := ioutil.ReadFile(filepath)
	if err != nil {
		return errors.Wrap(err, "failed to read file: "+filepath)
	}

	if err := yaml.Unmarshal(contents, c); err != nil {
		return errors.Wrap(err, "failed to parse configuration")
	}

	if err := c.Valid(); err != nil {
		return errors.Wrap(err, "config failed to pass validation")
	}

	return nil
}

func (c *Config) Valid() error {
	for _, w := range c.Workspaces {
		if w.AccessToken == "" {
			return errors.New("access token is required")
		}
	}
	return nil
}

func (w *Workspace) isInGroup(group string) bool {
	for _, g := range w.Groups {
		if string(g) == group {
			return true
		}
	}
	return false
}
