/*
Copyright 2021 The terraform-docs Authors.

Licensed under the MIT license (the "License"); you may not
use this file except in compliance with the License.

You may obtain a copy of the License at the LICENSE file in
the root directory of this source tree.
*/

package cli

import (
	"fmt"

	"github.com/terraform-docs/terraform-docs/internal/print"
	"github.com/terraform-docs/terraform-docs/internal/terraform"
)

type sections struct {
	Show    []string `yaml:"show"`
	Hide    []string `yaml:"hide"`
	ShowAll bool     `yaml:"show-all"`
	HideAll bool     `yaml:"hide-all"`

	header       bool `yaml:"-"`
	inputs       bool `yaml:"-"`
	modulecalls  bool `yaml:"-"`
	outputs      bool `yaml:"-"`
	providers    bool `yaml:"-"`
	requirements bool `yaml:"-"`
	resources    bool `yaml:"-"`
}

func defaultSections() sections {
	return sections{
		Show:    []string{},
		Hide:    []string{},
		ShowAll: true,
		HideAll: false,

		header:       false,
		inputs:       false,
		modulecalls:  false,
		outputs:      false,
		providers:    false,
		requirements: false,
		resources:    false,
	}
}

func (s *sections) validate() error {
	items := []string{"header", "inputs", "modules", "outputs", "providers", "requirements", "resources"}
	for _, item := range s.Show {
		switch item {
		case items[0], items[1], items[2], items[3], items[4], items[5], items[6]:
		default:
			return fmt.Errorf("'%s' is not a valid section", item)
		}
	}
	for _, item := range s.Hide {
		switch item {
		case items[0], items[1], items[2], items[3], items[4], items[5], items[6]:
		default:
			return fmt.Errorf("'%s' is not a valid section", item)
		}
	}
	if s.ShowAll && s.HideAll {
		return fmt.Errorf("'--show-all' and '--hide-all' can't be used together")
	}
	if s.ShowAll && len(s.Show) != 0 {
		return fmt.Errorf("'--show-all' and '--show' can't be used together")
	}
	if s.HideAll && len(s.Hide) != 0 {
		return fmt.Errorf("'--hide-all' and '--hide' can't be used together")
	}
	return nil
}

func (s *sections) visibility(section string) bool {
	if s.ShowAll && !s.HideAll {
		for _, n := range s.Hide {
			if n == section {
				return false
			}
		}
		return true
	}
	for _, n := range s.Show {
		if n == section {
			return true
		}
	}
	for _, n := range s.Hide {
		if n == section {
			return false
		}
	}
	return false
}

type outputvalues struct {
	Enabled bool   `yaml:"enabled"`
	From    string `yaml:"from"`
}

func defaultOutputValues() outputvalues {
	return outputvalues{
		Enabled: false,
		From:    "",
	}
}

func (o *outputvalues) validate() error {
	if o.Enabled && o.From == "" {
		if changedfs["output-values-from"] {
			return fmt.Errorf("value of '--output-values-from' can't be empty")
		}
		return fmt.Errorf("value of '--output-values-from' is missing")
	}
	return nil
}

type sortby struct {
	Required bool `name:"required"`
	Type     bool `name:"type"`
}
type sort struct {
	Enabled bool     `yaml:"enabled"`
	ByList  []string `yaml:"by"`
	By      sortby   `yaml:"-"`
}

func defaultSort() sort {
	return sort{
		Enabled: true,
		ByList:  []string{},
		By: sortby{
			Required: false,
			Type:     false,
		},
	}
}

func (s *sort) validate() error {
	types := []string{"required", "type"}
	for _, item := range s.ByList {
		switch item {
		case types[0], types[1]:
		default:
			return fmt.Errorf("'%s' is not a valid sort type", item)
		}
	}
	if s.By.Required && s.By.Type {
		return fmt.Errorf("'--sort-by-required' and '--sort-by-type' can't be used together")
	}
	return nil
}

type settings struct {
	Anchors   bool `yaml:"anchors"`
	Color     bool `yaml:"color"`
	Escape    bool `yaml:"escape"`
	Indent    int  `yaml:"indent"`
	Required  bool `yaml:"required"`
	Sensitive bool `yaml:"sensitive"`
}

func defaultSettings() settings {
	return settings{
		Anchors:   true,
		Color:     true,
		Escape:    true,
		Indent:    2,
		Required:  true,
		Sensitive: true,
	}
}

func (s *settings) validate() error {
	return nil
}

// Config represents all the available config options that can be accessed and passed through CLI
type Config struct {
	File         string       `yaml:"-"`
	Formatter    string       `yaml:"formatter"`
	HeaderFrom   string       `yaml:"header-from"`
	Sections     sections     `yaml:"sections"`
	OutputValues outputvalues `yaml:"output-values"`
	Sort         sort         `yaml:"sort"`
	Settings     settings     `yaml:"settings"`
}

// DefaultConfig returns new instance of Config with default values set
func DefaultConfig() *Config {
	return &Config{
		File:         "",
		Formatter:    "",
		HeaderFrom:   "main.tf",
		Sections:     defaultSections(),
		OutputValues: defaultOutputValues(),
		Sort:         defaultSort(),
		Settings:     defaultSettings(),
	}
}

// process provided Config
func (c *Config) process() {
	// sections
	if c.Sections.HideAll && !changedfs["show-all"] {
		c.Sections.ShowAll = false
	}
	if !c.Sections.ShowAll && !changedfs["hide-all"] {
		c.Sections.HideAll = true
	}
	c.Sections.header = c.Sections.visibility("header")
	c.Sections.inputs = c.Sections.visibility("inputs")
	c.Sections.modulecalls = c.Sections.visibility("modules")
	c.Sections.outputs = c.Sections.visibility("outputs")
	c.Sections.providers = c.Sections.visibility("providers")
	c.Sections.requirements = c.Sections.visibility("requirements")
	c.Sections.resources = c.Sections.visibility("resources")
}

// validate config and check for any misuse or misconfiguration
func (c *Config) validate() error {
	// formatter
	if c.Formatter == "" {
		return fmt.Errorf("value of 'formatter' can't be empty")
	}

	// header-from
	if c.HeaderFrom == "" {
		return fmt.Errorf("value of '--header-from' can't be empty")
	}

	// sections
	if err := c.Sections.validate(); err != nil {
		return err
	}

	// output values
	if err := c.OutputValues.validate(); err != nil {
		return err
	}

	// sort
	if err := c.Sort.validate(); err != nil {
		return err
	}

	// settings
	if err := c.Settings.validate(); err != nil {
		return err
	}

	return nil
}

// extract and build print.Settings and terraform.Options out of Config
func (c *Config) extract() (*print.Settings, *terraform.Options) {
	settings := print.DefaultSettings()
	options := terraform.NewOptions()

	// header-from
	options.HeaderFromFile = c.HeaderFrom

	// sections
	settings.ShowHeader = c.Sections.header
	settings.ShowInputs = c.Sections.inputs
	settings.ShowModuleCalls = c.Sections.modulecalls
	settings.ShowOutputs = c.Sections.outputs
	settings.ShowProviders = c.Sections.providers
	settings.ShowRequirements = c.Sections.requirements
	settings.ShowResources = c.Sections.resources
	options.ShowHeader = settings.ShowHeader

	// output values
	settings.OutputValues = c.OutputValues.Enabled
	options.OutputValues = c.OutputValues.Enabled
	options.OutputValuesPath = c.OutputValues.From

	// sort
	settings.SortByName = c.Sort.Enabled
	settings.SortByRequired = c.Sort.Enabled && c.Sort.By.Required
	settings.SortByType = c.Sort.Enabled && c.Sort.By.Type
	options.SortBy.Name = settings.SortByName
	options.SortBy.Required = settings.SortByRequired
	options.SortBy.Type = settings.SortByType

	// settings
	settings.EscapeCharacters = c.Settings.Escape
	settings.IndentLevel = c.Settings.Indent
	settings.ShowColor = c.Settings.Color
	settings.ShowRequired = c.Settings.Required
	settings.ShowSensitivity = c.Settings.Sensitive
	settings.ShowAnchors = c.Settings.Anchors

	return settings, options
}
