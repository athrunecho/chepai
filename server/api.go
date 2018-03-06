package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/athrunecho/chepai"
	"github.com/gin-gonic/gin"
	"github.com/northbright/jwthelper"
)

func getLoginID(c *gin.Context) (string, error) {
	cookie, err := c.Request.Cookie("jwt")
	switch err {
	case http.ErrNoCookie:
		return "", fmt.Errorf("no jwt found in cookie")
	case nil:
		tokenString := cookie.Value
		parser := jwthelper.NewRSASHAParser([]byte(rsaPubPEM))
		m, err := parser.Parse(tokenString)
		if err != nil {
			return "", fmt.Errorf("parser.Parse() error: %v", err)
		}

		// Convert interface{} to string
		ID, ok := m["id"].(string)
		if !ok {
			return "", fmt.Errorf("failed to convert interface{} to string")
		}
		return ID, nil
	default:
		return "", fmt.Errorf("get JWT cookie error: %v", err)
	}

}

func getUnixNanoTimeInfo(c *gin.Context) {
	var (
		err      error
		timeInfo chepai.TimeInfo
		reply    = chepai.UnixNanoTimeInfoReply{}
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getUnixNanoTimeInfo() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	timeInfo = cp.GetTimeInfo()
	reply.BeginTime = timeInfo.BeginTime.UnixNano()
	reply.PhaseOneEndTime = timeInfo.PhaseOneEndTime.UnixNano()
	reply.PhaseTwoEndTime = timeInfo.PhaseTwoEndTime.UnixNano()

	reply.Success = true

	//log.Printf("getUnixNanoTimeInfo() OK, reply: %v", reply)
}

func getUnixTimeInfo(c *gin.Context) {
	var (
		err      error
		timeInfo chepai.TimeInfo
		reply    = chepai.UnixTimeInfoReply{}
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getUnixTimeInfo() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	timeInfo = cp.GetTimeInfo()
	reply.BeginTime = timeInfo.BeginTime.Unix()
	reply.PhaseOneEndTime = timeInfo.PhaseOneEndTime.Unix()
	reply.PhaseTwoEndTime = timeInfo.PhaseTwoEndTime.Unix()

	reply.Success = true

	//log.Printf("getUnixTimeInfo() OK, reply: %v", reply)
}

func validLogin(ID, password string) bool {
	if ID == "" || password == "" {
		return false
	}

	if ID != password {
		return false
	}
	return true
}

func loginPOST(c *gin.Context) {
	type Req struct {
		ID       string `json:"id"`
		Password string `json:"password"`
	}

	var (
		err   error
		r     Req
		reply chepai.Reply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("LoginPOST() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if err = c.BindJSON(&r); err != nil {
		err = fmt.Errorf("invalid request")
		return
	}

	if !validLogin(r.ID, r.Password) {
		err = fmt.Errorf("incorrect password")
		return
	}

	signer := jwthelper.NewRSASHASigner([]byte(rsaPrivPEM))
	tokenString, err := signer.SignedString(
		jwthelper.NewClaim("id", r.ID),
	)
	if err != nil {
		return
	}
	cookie := jwthelper.NewCookie(tokenString)
	http.SetCookie(c.Writer, cookie)

	reply.Success = true
	//log.Printf("LoginPOST() OK: reply: %v", reply)
}

func getStartPrice(c *gin.Context) {
	var (
		err   error
		reply chepai.StartPriceReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getStartPrice() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		return
	}

	reply.StartPrice = cp.StartPrice
	reply.Success = true
	//log.Printf("getStartPrice() OK, %v", reply)
}

func getLicensePlateNum(c *gin.Context) {
	var (
		err   error
		reply chepai.LicensePlateNumReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getLicensePlateNum() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	reply.LicensePlateNum = cp.LicensePlateNum
	reply.Success = true
	//log.Printf("getLicensePlateNum() OK,  reply: %v", reply)
}

func getLowestPrice(c *gin.Context) {
	var (
		err   error
		phase int
		reply chepai.LowestPriceReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getLowestPrice() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	phase = cp.GetPhase(time.Now())
	if reply.LowestPrice, err = cp.ComupteLowestPrice(phase); err != nil {
		return
	}

	reply.Success = true
	//log.Printf("getLowestPrice() OK, reply: %v", reply)
}

func getBidderNum(c *gin.Context) {
	var (
		err   error
		reply chepai.BidderNumReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getBidderNum() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	if reply.BidderNum, err = cp.GetBidderNum(); err != nil {
		log.Printf("GetBidderNum() error: %v")
		return
	}

	reply.Success = true
	//log.Printf("getLowestPrice() OK, reply: %v", reply)
}

func bid(c *gin.Context) {
	type Req struct {
		Price int64 `json:"price"`
	}

	var (
		err   error
		r     Req
		reply chepai.BidReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("bid() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if err = c.BindJSON(&r); err != nil {
		err = fmt.Errorf("invalid request")
		return
	}

	reply.Price = r.Price

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	reply.Phase = cp.GetPhase(time.Now())

	if err = cp.Bid(reply.ID, reply.Price); err != nil {
		return
	}

	reply.Success = true
	//log.Printf("bid() OK: reply: %v", reply)
}

func getBidRecords(c *gin.Context) {
	var (
		err   error
		reply chepai.BidRecordsReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getBidRecords() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	if reply.Records, err = cp.GetBidRecordsByID(reply.ID); err != nil {
		log.Printf("GetBidRecordsByID() error: %v", err)
		return
	}

	reply.Success = true
	log.Printf("getBidRecords() OK, ID: %v:", reply.ID)
	for _, r := range reply.Records {
		log.Printf("%v: %v", r.Time, r.Price)
	}
}

func getResults(c *gin.Context) {
	var (
		err   error
		reply chepai.ResultsReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getResults() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	if reply.Results, err = cp.GetResults(); err != nil {
		log.Printf("GetResults() error: %v", err)
		return
	}

	reply.Success = true
	//log.Printf("getResults() OK, reply: %v", reply)
}

func getResult(c *gin.Context) {
	var (
		err   error
		reply chepai.ResultReply
	)

	defer func() {
		if err != nil {
			reply.ErrMsg = err.Error()
			log.Printf("getResult() error: %v", err)
		}

		c.JSON(200, reply)
	}()

	if reply.ID, err = getLoginID(c); err != nil {
		log.Printf("getLoginID() error")
		return
	}

	if reply.Done, reply.Price, err = cp.GetResultByID(reply.ID); err != nil {
		log.Printf("GetResultByID() error: %v", err)
		return
	}

	reply.Success = true
	//log.Printf("getResult() OK, reply: %v", reply)
}
