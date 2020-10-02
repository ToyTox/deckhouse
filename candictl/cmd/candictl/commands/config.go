package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
	"sigs.k8s.io/yaml"

	"flant/candictl/pkg/app"
	"flant/candictl/pkg/config"
	"flant/candictl/pkg/log"
	"flant/candictl/pkg/template"
)

const (
	bashibleTemplateOpenAPI = "/deckhouse/candi/bashible/openapi.yaml"
	kubeadmTemplateOpenAPI  = "/deckhouse/candi/control-plane-kubeadm/openapi.yaml"
)

func DefineRenderBashibleBundle(parent *kingpin.CmdClause) *kingpin.CmdClause {
	cmd := parent.Command("bashible-bundle", "Render bashible bundle.")
	app.DefineConfigFlags(cmd)
	app.DefineRenderConfigFlags(cmd)

	runFunc := func() error {
		templateData, err := config.ParseBashibleConfig(app.ConfigPath, bashibleTemplateOpenAPI)
		if err != nil {
			return err
		}

		templateController := template.NewTemplateController(app.RenderBashibleBundleDir)
		log.InfoF("Bundle Dir: %q\n\n", templateController.TmpDir)

		return template.PrepareBashibleBundle(
			templateController,
			templateData,
			strings.ToLower(templateData["provider"].(string)),
			templateData["bundle"].(string),
			"",
		)
	}

	cmd.Action(func(c *kingpin.ParseContext) error {
		err := log.Process("bootstrap", "Prepare Bashible Bundle", runFunc)
		if err != nil {
			log.ErrorF("\nCritical Error: %s\n", err)
			os.Exit(1)
		}
		return nil
	})

	return cmd
}

func DefineRenderKubeadmConfig(parent *kingpin.CmdClause) *kingpin.CmdClause {
	cmd := parent.Command("kubeadm-config", "Render kubeadm config.")
	app.DefineConfigFlags(cmd)
	app.DefineRenderConfigFlags(cmd)

	runFunc := func() error {
		templateData, err := config.ParseBashibleConfig(app.ConfigPath, kubeadmTemplateOpenAPI)
		if err != nil {
			return err
		}

		templateController := template.NewTemplateController(app.RenderBashibleBundleDir)
		log.InfoF("Bundle Dir: %q\n\n", templateController.TmpDir)

		return template.PrepareKubeadmConfig(templateController, templateData)
	}

	cmd.Action(func(c *kingpin.ParseContext) error {
		err := log.Process("bootstrap", "Prepare Kubeadm Config", runFunc)
		if err != nil {
			log.ErrorF("\nCritical Error: %s\n", err)
			os.Exit(1)
		}
		return nil
	})

	return cmd
}

func DefineCommandParseClusterConfiguration(kpApp *kingpin.Application, parentCmd *kingpin.CmdClause) *kingpin.CmdClause {
	var parseCmd *kingpin.CmdClause
	if parentCmd == nil {
		parseCmd = kpApp.Command("parse-cluster-configuration", "Parse configuration and print it.")
	} else {
		parseCmd = parentCmd.Command("cluster-configuration", "Parse configuration and print it.")
	}
	app.DefineInputOutputRenderFlags(parseCmd)

	parseCmd.Action(func(c *kingpin.ParseContext) error {
		var err error
		var metaConfig *config.MetaConfig

		// Should be fixed in kingpin repo or shell-operator and others should migrate to github.com/alecthomas/kingpin.
		// https://github.com/flant/kingpin/pull/1
		// replace gopkg.in/alecthomas/kingpin.v2 => github.com/flant/kingpin is not working
		if app.ParseInputFile == "" {
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read configs from stdin: %v", err)
			}
			metaConfig, err = config.ParseConfigFromData(string(data))
			if err != nil {
				return err
			}
		} else {
			metaConfig, err = config.ParseConfig(app.ParseInputFile)
			if err != nil {
				return err
			}
		}

		var output []byte
		switch app.ParseOutput {
		case "yaml":
			output, _ = yaml.Marshal(metaConfig)
		case "json":
			output = metaConfig.MarshalConfig()
		default:
			return fmt.Errorf("unknown output type: %s", app.ParseOutput)
		}
		fmt.Print(string(output))
		return nil
	})

	return parseCmd
}

func DefineCommandParseCloudDiscoveryData(kpApp *kingpin.Application, parentCmd *kingpin.CmdClause) *kingpin.CmdClause {
	var parseCmd *kingpin.CmdClause
	if parentCmd == nil {
		parseCmd = kpApp.Command("parse-cloud-discovery-data", "Parse cloud discovery data and print it.")
	} else {
		parseCmd = parentCmd.Command("cloud-discovery-data", "Parse cloud discovery data and print it.")
	}
	app.DefineInputOutputRenderFlags(parseCmd)

	parseCmd.Action(func(c *kingpin.ParseContext) error {
		var err error
		var data []byte

		if app.ParseInputFile == "" {
			data, err = ioutil.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("read cloud-discovery-data from stdin: %v", err)
			}
		} else {
			data, err = ioutil.ReadFile(app.ParseInputFile)
			if err != nil {
				return fmt.Errorf("loading input file: %v", err)
			}
		}

		schemaStore := config.NewSchemaStore()
		_, err = schemaStore.Validate(&data)
		if err != nil {
			return fmt.Errorf("validate cloud_discovery_data: %v", err)
		}

		var output []byte
		switch app.ParseOutput {
		case "yaml":
			output, _ = yaml.JSONToYAML(data)
		case "json":
			output = data
		default:
			return fmt.Errorf("unknown output type: %s", app.ParseOutput)
		}
		fmt.Print(string(output))
		return nil
	})

	return parseCmd
}