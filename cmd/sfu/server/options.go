package server

import (
	"errors"
	"os"
	"sync/atomic"
	"unsafe"

	alias "github.com/doublemo/baa/cores/conf"
)

// ConfigureOptions 配置文件服务
type ConfigureOptions struct {
	// 配置文件地址
	fp string

	// opts 配置信息
	opts unsafe.Pointer
}

// Read 加载配置文件
func (conf *ConfigureOptions) Read() *Config {
	return (*Config)(atomic.LoadPointer(&conf.opts))
}

// Load 加载配置文件
func (conf *ConfigureOptions) Load() error {
	if conf.fp == "" {
		return errors.New("config file does not exist")
	}

	if _, err := os.Stat(conf.fp); os.IsNotExist(err) {
		return errors.New("config file does not exist")
	}

	opts := Config{}
	if err := alias.BindWithConfFile(conf.fp, &opts, "alias", "mapstructure"); err != nil {
		return err
	}

	conf.Reset(&opts)
	return nil
}

// Reset 重置配置文件
func (conf *ConfigureOptions) Reset(opts *Config) {
	atomic.StorePointer(&conf.opts, unsafe.Pointer(opts))
}

// NewConfigureOptions 创建配置文件
func NewConfigureOptions(fp string, opts *Config) *ConfigureOptions {
	c := &ConfigureOptions{fp: fp}
	if opts == nil {
		opts = &Config{}
	}
	atomic.StorePointer(&c.opts, unsafe.Pointer(opts))
	return c
}
