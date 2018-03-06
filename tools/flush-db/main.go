package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/athrunecho/chepai"
	"github.com/garyburd/redigo/redis"
	"github.com/northbright/pathhelper"
	"github.com/northbright/redishelper"
)

// Config represents the app settings.
type Config struct {
	ServerAddr       string `json:"server_addr"`
	RedisServer      string `json:"redis_server"`
	RedisPassword    string `json:"redis_password"`
	StartAfter       int    `json:"start_after"`
	PhaseOneDuration int    `json:"phase_one_duration"`
	PhaseTwoDuration int    `json:"phase_two_duration"`
	StartPrice       int64  `json:"start_price"`
	LicensePlateNum  int64  `json:"license_plate_num"`
}

var (
	currentDir, configFile string
	config                 Config
	pool                   *redis.Pool
	cp                     *chepai.Chepai
)

func main() {
	var (
		err error
	)

	defer func() {
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}()

	if err = loadConfig(configFile, &config); err != nil {
		err = fmt.Errorf("loadConfig() error: %v", err)
		return
	}

	// New a redis pool
	pool = redishelper.NewRedisPool(":6379", "", 1000, 100, 60, true)
	defer pool.Close()

	// New a Chepai instance
	cp = chepai.New(pool,
		config.StartAfter,
		config.PhaseOneDuration,
		config.PhaseTwoDuration,
		config.StartPrice,
		config.LicensePlateNum,
	)

	if err = cp.FlushDB(); err != nil {
		fmt.Printf("flush db error: %v\n", err)
		return
	}

	fmt.Printf("flush db ok\n")
}

// init initializes path variables.
func init() {
	currentDir, _ = pathhelper.GetCurrentExecDir()
	configFile = path.Join(currentDir, "config.json")
}

// loadConfig loads app config.
func loadConfig(configFile string, config *Config) error {
	// Load Conifg
	buf, err := ioutil.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("load config file error: %v", err)

	}

	if err = json.Unmarshal(buf, config); err != nil {
		return fmt.Errorf("parse config err: %v", err)
	}

	return nil
}
