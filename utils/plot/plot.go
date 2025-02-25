package plot

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	dspOS "github.com/saveio/dsp-go-sdk/utils/os"
	"github.com/saveio/themis/common/log"
)

// executables are embedded
//go:embed bin/*
var s embed.FS

const (
	SYS_LINUX = "linux"
	SYS_WIN   = "win"

	FLAG_NUMERIC_ID  = "-i"
	FLAG_START_NONCE = "-s"
	FLAG_NONCES      = "-n"
	FLAG_PATH        = "-p"

	DEFAULT_PLOT_TOOL_NAME = "engraver_cpu"

	DEFAULT_PLOT_SIZEKB = 256
)

type PlotConfig struct {
	Sys        string // windows or linux
	NumericID  string // numeric ID
	StartNonce uint64 // start nonce
	Nonces     uint64 // num of nonce
	Path       string // path to store plot file
}

// do the file plotting with cfg, if cfg not exist, use saved cfg
func Plot(cfg *PlotConfig) error {

	if err := checkPlotConfig(cfg); err != nil {
		return err
	}

	toolName := getPlotToolName(cfg)
	toolPath := "./" + toolName

	if !fileExists(toolPath) {
		if err := loadPlotTool(cfg); err != nil {
			return fmt.Errorf("loadPlotTool error %s", err)
		}
	}

	if err := dspOS.CreateDirIfNeed(cfg.Path); err != nil {
		return fmt.Errorf("create dir in path %s errror %s", cfg.Path, err)
	}

	if err := runPlotCmd(cfg, toolPath); err != nil {
		log.Errorf("runPlotCmd error %s", err)
		return fmt.Errorf("runPlotCmd error %s", err)
	}

	log.Debugf("run plot cmd ok with config: %+v", cfg)
	return nil
}

func fileExists(filename string) bool {
	fi, err := os.Stat(filename)
	if fi != nil || (err != nil && !os.IsNotExist(err)) {
		return true
	}
	return false
}

func checkPlotConfig(cfg *PlotConfig) error {
	if cfg == nil {
		return fmt.Errorf("cfg is nil")
	}

	if cfg.Sys != SYS_LINUX && cfg.Sys != SYS_WIN {
		return fmt.Errorf("wrong sys %s", cfg.Sys)
	}

	_, err := strconv.Atoi(cfg.NumericID)
	if err != nil {
		return fmt.Errorf("invalid numeric id")
	}

	return nil
}

// the config should have been checked already
func runPlotCmd(cfg *PlotConfig, cmdPath string) error {
	var cmd *exec.Cmd

	if cfg.Path == "" {
		cmd = exec.Command(cmdPath, FLAG_NUMERIC_ID, cfg.NumericID, FLAG_START_NONCE, strconv.Itoa(int(cfg.StartNonce)),
			FLAG_NONCES, strconv.Itoa(int(cfg.Nonces)))
	} else {
		cmd = exec.Command(cmdPath, FLAG_NUMERIC_ID, cfg.NumericID, FLAG_START_NONCE, strconv.Itoa(int(cfg.StartNonce)),
			FLAG_NONCES, strconv.Itoa(int(cfg.Nonces)), FLAG_PATH, cfg.Path)
	}
	// dont check error here, it returns error even run successfully
	cmd.Run()

	fullPath := GetPlotFileFullPath(cfg)
	if !fileExists(fullPath) {
		return fmt.Errorf("plot file %s not found", fullPath)
	}
	return nil
}

func loadPlotTool(cfg *PlotConfig) error {
	log.Debugf("loadPlotTool from embed fs")

	data, err := s.ReadFile(getPlotToolFullPathFromEmbed(cfg))
	if err != nil {
		return fmt.Errorf("read ptool error %s", err)
	}

	fileName := "./" + getPlotToolName(cfg)
	err = ioutil.WriteFile(fileName, data, 0777)
	if err != nil {
		return fmt.Errorf("write tool file error %s", err)
	}
	return nil
}

func GetPlotFileName(cfg *PlotConfig) string {
	startStr := strconv.Itoa(int(cfg.StartNonce))
	// nonces should be multiplied by 8
	nonces := cfg.Nonces - cfg.Nonces%8
	noncesStr := strconv.Itoa(int(nonces))
	return strings.Join([]string{cfg.NumericID, startStr, noncesStr}, "_")
}

func GetPlotFileFullPath(cfg *PlotConfig) string {
	fileName := GetPlotFileName(cfg)
	return filepath.Join(cfg.Path, fileName)
}

func getPlotToolName(cfg *PlotConfig) string {
	name := DEFAULT_PLOT_TOOL_NAME
	if cfg.Sys == SYS_WIN {
		name += ".exe"
	}
	return name
}

func getPlotToolFullPathFromEmbed(cfg *PlotConfig) string {
	return "bin/" + cfg.Sys + "/" + getPlotToolName(cfg)
}
