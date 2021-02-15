package main

import (
	"flag"
	"os"

	"k8s.io/klog/v2"

	"bashible-apiserver/pkg/cmd/server"
	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/component-base/logs"
)

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	stopCh := genericapiserver.SetupSignalHandler()
	options := server.NewBashibleServerOptions(os.Stdout, os.Stderr)
	cmd := server.NewCommandStartBashibleServer(options, stopCh)
	cmd.Flags().AddGoFlagSet(flag.CommandLine)
	if err := cmd.Execute(); err != nil {
		klog.Fatal(err)
	}
}