package command

import (
	"fmt"
	"github.com/kenlabs/pando/pkg/system"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize server config file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkPandoRoot(); err != nil {
				return err
			}

			configFile := filepath.Join(Opt.PandoRoot, Opt.ConfigFile)
			fmt.Println("Init pando-server configs at ", configFile)
			if err := checkConfigExists(configFile); err != nil {
				return err
			}

			if err := setBandwidth(); err != nil {
				return err
			}

			if err := setIdentity(); err != nil {
				return err
			}

			if err := saveConfig(configFile); err != nil {
				return err
			}

			fmt.Printf("init complete.\n")
			return nil
		},
	}
}

func checkPandoRoot() error {
	const failedError = "check pando root failed:\n\t%v\n"
	rootExists, err := system.IsDirExists(Opt.PandoRoot)
	if err != nil {
		return fmt.Errorf(failedError, err)
	}
	if !rootExists {
		fmt.Printf("pando root %s does not exist, try to create...\n", Opt.PandoRoot)
		err := os.MkdirAll(Opt.PandoRoot, 0755)
		if err != nil {
			return fmt.Errorf("create pando root %s failed: %v", Opt.PandoRoot, err)
		}
	}

	rootWritable, err := system.IsDirWritable(Opt.PandoRoot)
	if err != nil {
		return fmt.Errorf(failedError, err)
	}
	if !rootWritable {
		return fmt.Errorf("pando root %s is not writable\n", Opt.PandoRoot)
	}

	return nil
}

func checkConfigExists(configFile string) error {
	configExists, err := system.IsFileExists(configFile)
	if err != nil {
		return fmt.Errorf("init config failed: %v", err)
	}
	if configExists {
		return fmt.Errorf("config file exists: %s", configFile)
	}
	return nil
}

func setBandwidth() error {
	var err error
	if !Opt.DisableSpeedTest {
		Opt.RateLimit.Bandwidth, err = system.TestInternetSpeed(false)
		if err != nil {
			return err
		}
	} else {
		Opt.RateLimit.Bandwidth = 10.0
	}
	return nil
}

func setIdentity() error {
	var err error
	Opt.Identity.PeerID, Opt.Identity.PrivateKey, err = system.CreateIdentity()
	if err != nil {
		return err
	}
	return nil
}

func saveConfig(configFile string) error {
	buff, err := yaml.Marshal(Opt)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(configFile, os.O_RDWR|os.O_CREATE, 0755)
	defer func(file *os.File) {
		_ = file.Close()
	}(file)
	if err != nil {
		return err
	}

	_, err = file.WriteString(string(buff))
	if err != nil {
		return err
	}

	return nil
}
