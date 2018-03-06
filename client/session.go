package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
	//	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/athrunecho/chepai"
)

// NewSession creates a new session of ming800.
func NewSession(serverURLStr string) (*Session, error) {
	var err error

	s := &Session{ServerURLStr: serverURLStr}

	if s.ServerURL, err = url.Parse(s.ServerURLStr); err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	s.client = &http.Client{Jar: jar}
	return s, err
}

// Login performs the login action.
func (s *Session) Login(ID, password string) error {
	var reply chepai.Reply
	// Login.
	data := struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}{ID, password}

	buf, _ := json.Marshal(data)

	refURL, _ := url.Parse("/api/login")
	urlStr := s.ServerURL.ResolveReference(refURL).String()

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(buf))
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Store JWT cookie
	respCookies := resp.Cookies()
	if len(respCookies) != 1 || respCookies[0].Name != "jwt" {
		return fmt.Errorf("failed to get jwt in response cookies")
	}
	// Set cookie for cookiejar manually.
	s.client.Jar.SetCookies(s.ServerURL, respCookies)

	// Get reply
	if buf, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	if err = json.Unmarshal(buf, &reply); err != nil {
		return err
	}

	if !reply.Success {
		return fmt.Errorf(reply.ErrMsg)
	}
	return nil
}

func (s *Session) GetUnixNanoTimeInfo() (*chepai.TimeInfo, error) {
	var reply chepai.UnixNanoTimeInfoReply

	refURL, _ := url.Parse("/api/unix_nano_time_info")
	urlStr := s.ServerURL.ResolveReference(refURL).String()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(buf, &reply); err != nil {
		return nil, err
	}

	if !reply.Success {
		return nil, fmt.Errorf(reply.ErrMsg)
	}
	return &chepai.TimeInfo{
		time.Unix(0, reply.BeginTime),
		time.Unix(0, reply.PhaseOneEndTime),
		time.Unix(0, reply.PhaseTwoEndTime),
	}, nil
}

func (s *Session) GetStartPrice() (int64, error) {
	var reply chepai.StartPriceReply

	refURL, _ := url.Parse("/api/start_price")
	urlStr := s.ServerURL.ResolveReference(refURL).String()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if err = json.Unmarshal(buf, &reply); err != nil {
		return 0, err
	}

	if !reply.Success {
		return 0, fmt.Errorf(reply.ErrMsg)
	}
	return reply.StartPrice, nil
}

func (s *Session) GetLowestPrice() (int64, error) {
	var reply chepai.LowestPriceReply

	refURL, _ := url.Parse("/api/lowest_price")
	urlStr := s.ServerURL.ResolveReference(refURL).String()

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return 0, err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	if err = json.Unmarshal(buf, &reply); err != nil {
		return 0, err
	}

	if !reply.Success {
		return 0, fmt.Errorf(reply.ErrMsg)
	}
	return reply.LowestPrice, nil
}

func (s *Session) Bid(price int64) error {
	var reply chepai.BidReply

	data := struct {
		Price int64 `json:"price"`
	}{price}

	buf, _ := json.Marshal(data)

	refURL, _ := url.Parse("/api/bid")
	urlStr := s.ServerURL.ResolveReference(refURL).String()

	req, err := http.NewRequest("POST", urlStr, bytes.NewReader(buf))
	if err != nil {
		return err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Get reply
	if buf, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}

	if err = json.Unmarshal(buf, &reply); err != nil {
		return err
	}

	if !reply.Success {
		return fmt.Errorf(reply.ErrMsg)
	}
	return nil
}
