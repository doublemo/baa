// Copyright (c) 2019 The balala Authors <https://github.com/doublemo/baa>
//
// 编译方式,需要通过这种编译写版本信息
// VERSION = "0.0.1"
// COMMIT = $(shell git rev-parse HEAD) # --short
// BUILDDATE = $(shell date +%F@%T)
// go install -v -ldflags "-X main.version=$(VERSION) -X main.commitid=$(COMMIT) -X main.builddate=$(BUILDDATE)"
// go build -race -ldflags "-X main.version=$(VERSION) -X main.commitid=$(COMMIT) -X main.builddate=$(BUILDDATE)"
// GOOS=linux GOARCH=amd64 go install -ldflags "-X main.version=$(VERSION) -X main.commitid=$(COMMIT) -X main.builddate=$(BUILDDATE)"

package main

import (
	"flag"
	"fmt"
	"os"

	coresos "github.com/doublemo/baa/cores/os"
	"github.com/doublemo/baa/kits/kun"
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
Usage: agent-server [options]
Server Options:
	-c, --config <file>              Configuration file

Windows Services:
		--install                    Install this server to Windows Services 
		--uninstall                  Uninstall this server in Windows Services 
		--dname                      The name displayed in the windows service 
		--description                Description displayed in Windows Services
		--args                       Parameters running in Windows Service
	
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
	fmt.Printf("agent-server version %s commitid %s builddate %s\n", version, commitid, builddate)
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

		// install this server to Windows Services
		install bool

		// uninstall this server in Windows Services
		uninstall bool

		// dname The name displayed in the windows service
		dname string

		// description displayed in Windows Services
		description string

		// args Parameters running in Windows Service
		args string
	)

	fs := flag.NewFlagSet("agent", flag.ExitOnError)
	fs.Usage = usage
	fs.BoolVar(&showHelp, "h", false, "Show this message.")
	fs.BoolVar(&showHelp, "help", false, "Show this message.")
	fs.StringVar(&fp, "c", "conf/kun.conf", "Configuration file")
	fs.StringVar(&fp, "config", "conf/kun.conf", "Configuration file")
	fs.BoolVar(&showVersion, "version", false, "Print version information.")
	fs.BoolVar(&showVersion, "v", false, "Print version information.")
	fs.BoolVar(&install, "install", false, "Install this server to Windows Services")
	fs.BoolVar(&uninstall, "uninstall", false, "Uninstall this server in Windows Services")
	fs.StringVar(&dname, "dname", "Balala Agent", "The name displayed in the windows service")
	fs.StringVar(&description, "description", "Balala agent server", "Description displayed in Windows Services")
	fs.StringVar(&args, "args", "", "Parameters running in Windows Service")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if showHelp {
		usage()
	}

	if showVersion {
		ver()
	}

	// if install {
	// 	installService(service.Name, dname, description, args)
	// }

	// if uninstall {
	// 	uninstallService(service.Name)
	// }

	opts := kun.NewConfigureOptions(fp, nil)
	if err := opts.Load(); err != nil {
		panic(err)
	}

	if err := coresos.Run(kun.New(opts)); err != nil {
		panic(err)
	}
}
