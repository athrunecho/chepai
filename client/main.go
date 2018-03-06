package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/northbright/pathhelper"
)

// Config represents the app settings.
type Config struct {
	ServerAddr string `json:"server_addr"`
	ClientName string `json:"client_name"`
	BidderNum  int64  `json:"bidder_num"`
}

var (
	currentDir, configFile string
	config                 Config
)

type Session struct {
	ServerURLStr string
	ServerURL    *url.URL
	client       *http.Client
}

var ()

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

	Emu(config)
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

func Emu(config Config) {
	sem := make(chan struct{}, config.BidderNum)

	for i := int64(0); i < config.BidderNum; i++ {
		// After first "concurrency" amount of goroutines started,
		// It'll block starting new goroutines until one running goroutine finishs.
		sem <- struct{}{}
		go EmuBid(config, sem, i)

	}

	// After last goroutine is started,
	// there're still "concurrency" amount of goroutines running.
	// Make sure wait all goroutines to finish.
	for j := 0; j < cap(sem); j++ {
		sem <- struct{}{}
		//log.Printf("----- j: %v\n", j)
	}
}

func EmuBid(config Config, sem chan struct{}, i int64) {
	defer func() { <-sem }()
	// New session
	s, err := NewSession(config.ServerAddr)
	if err != nil {
		log.Printf("NewSession() error: %v", err)
		return
	}

	// Login
	ID := strconv.FormatInt(i, 10)
	ID = fmt.Sprintf("%s:%s", config.ClientName, ID)
	if err = s.Login(ID, ID); err != nil {
		log.Printf("Login() error: %v", err)
		return
	}

	// Get time info
	info, err := s.GetUnixNanoTimeInfo()
	if err != nil {
		log.Printf("GetUnixNanoTimeInfo() error: %v", err)
		return
	}

	// Get start price
	startPrice, err := s.GetStartPrice()
	if err != nil {
		log.Printf("GetStartPrice() error: %v", err)
		return
	}

	// Generate sleep duration before phase one end
	duration, err := genSleepTime(ID, info.BeginTime, info.PhaseOneEndTime.Add((-3)*time.Second))
	if err != nil {
		log.Printf("gen phase one sleep time error: %v", err)
		return
	}
	time.Sleep(duration)

	//log.Printf("ID: %v, phase 1 wake: %v", ID, time.Now())
	if err = s.Bid(startPrice); err != nil {
		log.Printf("ID: %v, bid on phase one error: %v", ID, err)
		return
	}

	// Generate sleep duration before phase two end
	duration, err = genSleepTime(ID, info.PhaseOneEndTime, info.PhaseTwoEndTime.Add((-3)*time.Second))
	if err != nil {
		log.Printf("gen phase two sleep time error: %v", err)
		return
	}
	time.Sleep(duration)

	//log.Printf("ID: %v, phase 2 wake: %v", ID, time.Now())
	lowestPrice, err := s.GetLowestPrice()
	if err != nil {
		log.Printf("get lowest price on phase two error: %v", err)
		return
	}

	// Generate bid price for phase two
	price := generatePhaseTwoPrice(lowestPrice)
	if err = s.Bid(price); err != nil {
		log.Printf("ID: %v, bid on phase two error: %v", ID, err)
		return
	}

	//log.Printf("i: %v\n", i)
}
