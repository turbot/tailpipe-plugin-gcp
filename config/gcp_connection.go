package config

import "github.com/turbot/tailpipe-plugin-sdk/parse"

type GcpConnection struct {
}

func NewGcpConnection() parse.Config {
	return &GcpConnection{}
}

func (c *GcpConnection) Validate() error {
	return nil
}
