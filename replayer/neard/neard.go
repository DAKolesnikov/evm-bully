// Package neard implements NEAR daemon related functionality.
package neard

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/aurora-is-near/evm-bully/util/git"
	"github.com/ethereum/go-ethereum/log"
	"github.com/frankbraun/codechain/util/file"
)

type NEARDaemon struct {
	head string
}

func build(release bool) error {
	args := []string{
		"build",
		"--package", "neard",
		"--features", "protocol_feature_evm,nightly_protocol_features",
	}
	if release {
		args = append(args, "--release")
	}
	cmd := exec.Command("cargo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func initDaemon(release bool, localDir string) error {
	var name string
	if release {
		name = filepath.Join(".", "target", "release", "neard")
	} else {
		name = filepath.Join(".", "target", "debug", "neard")
	}
	cmd := exec.Command(name, "--home="+localDir, "--verbose=true", "init")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func editGenesis(localDir string) error {
	filename := filepath.Join(localDir, "genesis.json")
	backup := filepath.Join(localDir, "genesis_old.json")
	if err := file.Copy(filename, backup); err != nil {
		return err
	}
	// read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	// unmarshal
	jsn := make(map[string]interface{})
	if err := json.Unmarshal(data, &jsn); err != nil {
		return err
	}
	// "runtime_config" -> "wasm_config" -> "limit_config"
	runtimeConfig := jsn["runtime_config"].(map[string]interface{})
	wasmConfig := runtimeConfig["wasm_config"].(map[string]interface{})
	limitConfig := wasmConfig["limit_config"].(map[string]interface{})
	// change default values
	limitConfig["max_gas_burnt"] = 800000000000000
	limitConfig["max_total_prepaid_gas"] = 800000000000000
	// marshal
	data, err = json.MarshalIndent(jsn, "", "  ")
	if err != nil {
		return err
	}
	// write file
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}
	return nil
}

func Setup(release bool) (*NEARDaemon, error) {
	var n NEARDaemon
	log.Info("setup neard")
	// get cwd
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	// switch to nearcore directory
	nearDir := filepath.Join(cwd, "..", "nearcore")
	if err := os.Chdir(nearDir); err != nil {
		return nil, err
	}
	// get current HEAD
	n.head, err = git.Head()
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("head=%s", n.head))
	// backup .near/local, if it exists
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	localDir := filepath.Join(home, ".near", "local")
	exists, err := file.Exists(localDir)
	if err != nil {
		return nil, err
	}
	if exists {
		localOld := localDir + "_old"
		log.Info(fmt.Sprintf("mv %s %s", localDir, localOld))
		// remove old backup directory
		if err := os.RemoveAll(localOld); err != nil {
			return nil, err
		}
		// move
		if err := os.Rename(localDir, localOld); err != nil {
			return nil, err
		}
	} else {
		log.Info(fmt.Sprintf("directory '%s' does not exist", localDir))
	}
	// make sure neard is build
	if err := build(release); err != nil {
		return nil, err
	}
	// initialize neard
	if err := initDaemon(release, localDir); err != nil {
		return nil, err
	}
	// edit genesis.json
	if err := editGenesis(localDir); err != nil {
		return nil, err
	}
	// switch back to original directory
	if err := os.Chdir(cwd); err != nil {
		return nil, err
	}
	return &n, nil
}

func (n *NEARDaemon) Stop() {
	log.Info("stop neard")
}
