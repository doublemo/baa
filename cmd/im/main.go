// Copyright (c) 2019 The baa Authors <https://github.com/doublemo/baa>
//
// 编译方式,需要通过这种编译写版本信息
// VERSION = "0.0.1"
// COMMIT = $(shell git rev-parse HEAD) # --short
// BUILDDATE = $(shell date +%F@%T)
// go install -v -ldflags "-X main.version=$(VERSION) -X main.commitid=$(COMMIT) -X main.builddate=$(BUILDDATE)"
// go build -race -ldflags "-X main.version=$(VERSION) -X main.commitid=$(COMMIT) -X main.builddate=$(BUILDDATE)"
// GOOS=linux GOARCH=amd64 go install -ldflags "-X main.version=$(VERSION) -X main.commitid=$(COMMIT) -X main.builddate=$(BUILDDATE)"
// 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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/doublemo/baa/cmd/im/server"
	coresos "github.com/doublemo/baa/cores/os"
)

// 定义版本信息
var (
	// version 版本号
	version string

	// commitid 代码提交版本号
	commitid string

	// builddate 编译日期
	builddate string
)

var usageStr = `
Usage: im server [options]
Server Options:
	-c, --config <file>              Configuration file
	
Common Options:
    -h, --help                       Show this message
    -v, --version                    Show version
`

// usage will print out the flag options for the server.
func usage() {
	fmt.Printf("%s\n", usageStr)
	os.Exit(0)
}

func ver() {
	fmt.Printf("im server version %s commitid %s builddate %s\n", version, commitid, builddate)
	os.Exit(0)
}

func main() {
	var (
		// fp 配置文件地址
		fp string

		// showVersion 显示版本信息
		showVersion bool

		// showHelp 显示配置信息
		showHelp bool
	)

	fs := flag.NewFlagSet("im", flag.ExitOnError)
	fs.Usage = usage
	fs.BoolVar(&showHelp, "h", false, "Show this message.")
	fs.BoolVar(&showHelp, "help", false, "Show this message.")
	fs.StringVar(&fp, "c", "conf/im.conf", "Configuration file")
	fs.StringVar(&fp, "config", "conf/im.conf", "Configuration file")
	fs.BoolVar(&showVersion, "version", false, "Print version information.")
	fs.BoolVar(&showVersion, "v", false, "Print version information.")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if showHelp {
		usage()
	}

	if showVersion {
		ver()
	}

	opts := server.NewConfigureOptions(fp, nil)
	if err := opts.Load(); err != nil {
		panic(err)
	}

	if err := coresos.Run(server.New(opts)); err != nil {
		panic(err)
	}
}
