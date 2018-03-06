package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/athrunecho/chepai"
	"github.com/garyburd/redigo/redis"
	"github.com/gin-gonic/gin"
	"github.com/northbright/pathhelper"
	"github.com/northbright/redishelper"
)

// Config represents the app settings.
type Config struct {
	ServerAddr       string `json:"server_addr"`
	RedisServer      string `json:"redis_server"`
	RedisPassword    string `json:"redis_password"`
	RedisPoolSize    int    `json:"redis_pool_size"`
	StartAfter       int    `json:"start_after"`
	PhaseOneDuration int    `json:"phase_one_duration"`
	PhaseTwoDuration int    `json:"phase_two_duration"`
	StartPrice       int64  `json:"start_price"`
	LicensePlateNum  int64  `json:"license_plate_num"`
}

var (
	currentDir, configFile    string
	templatesPath, staticPath string
	config                    Config
	pool                      *redis.Pool
	cp                        *chepai.Chepai
)

func main() {
	var (
		err error
	)

	defer func() {
		if err != nil {
			log.Printf("error: %v\n", err)
		}
	}()

	if err = loadConfig(configFile, &config); err != nil {
		err = fmt.Errorf("loadConfig() error: %v", err)
		return
	}

	// New a redis pool
	pool = redishelper.NewRedisPool(":6379", "", config.RedisPoolSize, 100, 60, true)
	defer pool.Close()

	// New a Chepai instance
	cp = chepai.New(pool,
		config.StartAfter,
		config.PhaseOneDuration,
		config.PhaseTwoDuration,
		config.StartPrice,
		config.LicensePlateNum,
	)

	log.Printf("cp: %v", cp)
	// Flush DB before new chepai
	if err = cp.FlushDB(); err != nil {
		log.Printf("FlushDB() error: %v")
		return
	}

	r := gin.Default()

	// Serve Static files.
	r.Static("/static/", staticPath)

	// Load Templates.
	r.LoadHTMLGlob(fmt.Sprintf("%v/*", templatesPath))

	// Core APIs.
	r.POST("/api/login", loginPOST)
	r.GET("/api/unix_nano_time_info", getUnixNanoTimeInfo)
	r.GET("/api/unix_time_info", getUnixTimeInfo)
	r.GET("/api/start_price", getStartPrice)
	r.GET("/api/license_plate_num", getLicensePlateNum)
	r.GET("/api/lowest_price", getLowestPrice)
	r.GET("/api/bidder_num", getBidderNum)
	r.POST("/api/bid", bid)
	r.GET("/api/bid_records", getBidRecords)
	r.GET("/api/results", getResults)
	r.GET("/api/result", getResult)

	// Pages.
	r.GET("/", home)

	r.Run(config.ServerAddr)
}

// init initializes path variables.
func init() {
	currentDir, _ = pathhelper.GetCurrentExecDir()
	configFile = path.Join(currentDir, "config.json")
	templatesPath = path.Join(currentDir, "templates")
	staticPath = path.Join(currentDir, "static")
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
