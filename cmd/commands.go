package main

import (
	"fmt"
	"runtime/debug"

	"my-apps.com/myapp/internal/controller/config"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctlrZap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func createRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "gateway",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	return rootCmd
}

func createControllerCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "controller",
		Short: "Run the NGINX Gateway Fabric control plane",
		RunE: func(cmd *cobra.Command, _ []string) error {
			atom := zap.NewAtomicLevel()

			logger := ctlrZap.New(ctlrZap.Level(atom))
			klog.SetLogger(logger)

			commit, date, dirty := getBuildInfo()
			logger.Info(
				"Starting the NGINX Gateway Fabric control plane",
				"version", version,
				"commit", commit,
				"date", date,
				"dirty", dirty,
			)
			log.SetLogger(logger)

			conf := config.Config{
				GatewayCtlrName:  gatewayCtlrName.value,
				ConfigName:       configName.String(),
				Logger:           logger,
				GatewayClassName: gatewayClassName.value,
			}

			if err := controller.StartManager(conf); err != nil {
				return fmt.Errorf("failed to start control loop: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func getBuildInfo() (commitHash string, commitTime string, dirtyBuild string) {
	commitHash = "unknown"
	commitTime = "unknown"
	dirtyBuild = "unknown"

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}
	for _, kv := range info.Settings {
		switch kv.Key {
		case "vcs.revision":
			commitHash = kv.Value
		case "vcs.time":
			commitTime = kv.Value
		case "vcs.modified":
			dirtyBuild = kv.Value
		}
	}

	return
}
