package chepai

import (
	"fmt"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

type TimeInfo struct {
	BeginTime       time.Time
	PhaseOneEndTime time.Time
	PhaseTwoEndTime time.Time
}

type Chepai struct {
	pool            *redis.Pool
	timeInfo        TimeInfo
	StartPrice      int64
	LicensePlateNum int64
	mux             sync.Mutex
}

type BidRecord struct {
	Price int64 `redis:"price" json:"price"`
	Time  int64 `redis:"time" json:"time"`
}

func New(pool *redis.Pool, startAfter, phaseOneDuration, phaseTwoDuration int, startPrice, licensePlateNum int64) *Chepai {
	beginTime := time.Now().Add(time.Duration(startAfter) * time.Second)
	phaseOneEndTime := beginTime.Add(time.Duration(phaseOneDuration) * time.Second)
	phaseTwoEndTime := phaseOneEndTime.Add(time.Duration(phaseTwoDuration) * time.Second)

	return &Chepai{
		pool: pool,
		timeInfo: TimeInfo{
			BeginTime:       beginTime,
			PhaseOneEndTime: phaseOneEndTime,
			PhaseTwoEndTime: phaseTwoEndTime,
		},
		StartPrice:      startPrice,
		LicensePlateNum: licensePlateNum,
		mux:             sync.Mutex{},
	}
}

func (cp *Chepai) GetTimeInfo() TimeInfo {
	return cp.timeInfo
}

func (cp *Chepai) GetPhase(t time.Time) int {
	switch {
	case t.Before(cp.timeInfo.BeginTime):
		return 0
	case t.Equal(cp.timeInfo.BeginTime) || (t.After(cp.timeInfo.BeginTime) && t.Before(cp.timeInfo.PhaseOneEndTime)):
		return 1
	case t.Equal(cp.timeInfo.PhaseOneEndTime) || (t.After(cp.timeInfo.PhaseOneEndTime) && t.Before(cp.timeInfo.PhaseTwoEndTime)):
		return 2
	default:
		return 3
	}
}

func (cp *Chepai) ComputePhaseTwoLowestPrice() (int64, error) {
	conn := cp.pool.Get()
	defer conn.Close()

	k := "prices"
	prices, err := redis.Int64s(conn.Do("ZREVRANGE", k, 0, -1))
	if err != nil {
		return 0, err
	}

	sum := int64(0)
	for _, price := range prices {
		k := fmt.Sprintf("%v:num", price)
		num, err := redis.Int64(conn.Do("GET", k))
		if err != nil && err != redis.ErrNil {
			return 0, err
		}

		if num == 0 {
			return 0, fmt.Errorf("no member matches price: %v", price)
		}

		sum += num
		if sum >= cp.LicensePlateNum {
			return price, nil
		}
	}

	l := len(prices)
	if l > 0 {
		if prices[l-1] > cp.StartPrice {
			return cp.StartPrice, nil
		}
		return prices[l-1], nil
	}
	return cp.StartPrice, nil
}

func (cp *Chepai) ComupteLowestPrice(phase int) (int64, error) {
	switch phase {
	case 0, 1:
		return cp.StartPrice, nil
	case 2, 3:
		return cp.ComputePhaseTwoLowestPrice()
	default:
		return 0, fmt.Errorf("incorrect phase: %v", phase)
	}
}

func (cp *Chepai) ValidPhaseOnePrice(price int64) bool {
	if price != cp.StartPrice {
		return false
	}
	return true
}

func (cp *Chepai) ValidPhaseTwoPrice(lowestPrice, price int64) bool {
	diffArr := []int64{-300, -200, -100, 0, 100, 200, 300}

	diff := price - lowestPrice
	for _, v := range diffArr {
		if diff == v {
			return true
		}
	}
	return false
}

func (cp *Chepai) Bid(ID string, price int64) error {
	cp.mux.Lock()
	defer cp.mux.Unlock()

	t := time.Now()
	phase := cp.GetPhase(t)

	switch phase {
	case 1:
		if !cp.ValidPhaseOnePrice(price) {
			return fmt.Errorf("invalid phase one price: %v", price)
		}

		conn := cp.pool.Get()
		defer conn.Close()

		k := fmt.Sprintf("record:%v:phase:1", ID)
		exists, err := redis.Bool(conn.Do("EXISTS", k))
		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("already bid on phase one")
		}

		conn.Do("MULTI")
		conn.Send("HMSET", k, "time", t.UnixNano(), "price", price)
		conn.Send("INCR", "bidder_num")
		if _, err = conn.Do("EXEC"); err != nil {
			return err
		}
		return nil
	case 2:
		lowestPrice, err := cp.ComputePhaseTwoLowestPrice()
		if err != nil {
			return err
		}

		if !cp.ValidPhaseTwoPrice(lowestPrice, price) {
			return fmt.Errorf("invalid phase 2 price:%v", price)
		}

		conn := cp.pool.Get()
		defer conn.Close()

		k := fmt.Sprintf("record:%v:phase:1", ID)
		exists, err := redis.Bool(conn.Do("EXISTS", k))
		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("no bid record on phase one")
		}

		k = fmt.Sprintf("record:%v:phase:2", ID)
		exists, err = redis.Bool(conn.Do("EXISTS", k))
		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("already bid on phase two")
		}

		conn.Do("MULTI")
		conn.Send("HMSET", k, "time", t.UnixNano(), "price", price)
		conn.Send("ZADD", "prices", price, price)

		k = fmt.Sprintf("%v:num", price)
		conn.Send("INCR", k)

		k = fmt.Sprintf("%v:ids", price)
		conn.Send("ZADD", k, t.UnixNano(), ID)

		if _, err = conn.Do("EXEC"); err != nil {
			return err
		}
		return nil

	default:
		return fmt.Errorf("incorrect phase: %v", phase)
	}

	return nil
}

func (cp *Chepai) GetBidderNum() (int64, error) {
	conn := cp.pool.Get()
	defer conn.Close()

	num, err := redis.Int64(conn.Do("GET", "bidder_num"))
	if err != nil && err != redis.ErrNil {
		return 0, err
	}
	return num, nil
}

func (cp *Chepai) FlushDB() error {
	conn := cp.pool.Get()
	defer conn.Close()

	_, err := conn.Do("FLUSHDB")
	if err != nil {
		return err
	}
	return nil
}

func (cp *Chepai) GetBidRecordsByID(ID string) ([]*BidRecord, error) {
	var records []*BidRecord

	phase := cp.GetPhase(time.Now())

	conn := cp.pool.Get()
	defer conn.Close()

	switch phase {
	case 0:
		return nil, nil
	case 1:
		record, err := cp.getBidRecordByID(conn, phase, ID)
		if err != nil {
			return nil, err
		}
		if record != nil {
			records = append(records, record)
		}
		return records, nil

	case 2, 3:
		phases := []int{1, 2}
		for i := 0; i < len(phases); i++ {
			record, err := cp.getBidRecordByID(conn, phases[i], ID)
			if err != nil {
				return nil, err
			}

			if record != nil {
				records = append(records, record)
			}
		}
		return records, nil
	default:
		return nil, nil
	}
}

func (cp *Chepai) getBidRecordByID(conn redis.Conn, phase int, ID string) (*BidRecord, error) {
	var record BidRecord

	k := fmt.Sprintf("record:%v:phase:%v", ID, phase)
	exists, err := redis.Bool(conn.Do("EXISTS", k))
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, nil
	}

	values, err := redis.Values(conn.Do("HGETALL", k))
	if err != nil {
		return nil, err
	}

	if err = redis.ScanStruct(values, &record); err != nil {
		return nil, err
	}
	return &record, nil
}

func (cp *Chepai) GenerateResults() error {
	conn := cp.pool.Get()
	defer conn.Close()

	pipedConn := cp.pool.Get()
	defer pipedConn.Close()

	// Clear previous results
	_, err := conn.Do("DEL", "results")
	if err != nil {
		return err
	}

	k := "prices"
	prices, err := redis.Int64s(conn.Do("ZREVRANGE", k, 0, -1))
	if err != nil {
		return err
	}

	pipedConn.Do("MULTI")
	availableNum := cp.LicensePlateNum
	for _, price := range prices {
		if availableNum == 0 {
			break
		}

		k := fmt.Sprintf("%v:num", price)
		num, err := redis.Int64(conn.Do("GET", k))
		if err != nil && err != redis.ErrNil {
			return err
		}

		if num == 0 {
			return fmt.Errorf("no member matches price: %v", price)
		}

		k = fmt.Sprintf("%v:ids", price)
		stop := int64(-1)
		// all IDs should be in results
		if num < availableNum {
			availableNum -= num
		} else { // only first available num
			stop = availableNum - 1
			availableNum = 0
		}

		IDs, err := redis.Strings(conn.Do("ZRANGE", k, 0, stop))
		if err != nil {
			return err
		}

		// Append result
		for _, ID := range IDs {
			pipedConn.Send("HSET", "results", ID, price)
		}
	}

	if _, err = pipedConn.Do("EXEC"); err != nil {
		return err
	}
	return nil
}

func (cp *Chepai) GetResults() (map[string]string, error) {
	conn := cp.pool.Get()
	defer conn.Close()

	items := []string{}
	results := map[string]string{}
	cursor := 0

	for {
		v, err := redis.Values(conn.Do("HSCAN", "results", cursor, "COUNT", 1024))
		if err != nil {
			return map[string]string{}, err
		}

		if v, err = redis.Scan(v, &cursor, &items); err != nil {
			return map[string]string{}, err
		}

		l := len(items)
		if l > 0 {
			if l%2 != 0 {
				return map[string]string{}, fmt.Errorf("GetResults() error: HSCAN result error.")
			}

			for i := 0; i < l; i += 2 {
				results[items[i]] = items[i+1]
			}
		}

		if cursor == 0 {
			break
		}
	}
	return results, nil
}

func (cp *Chepai) GetResultByID(ID string) (bool, int64, error) {
	phase := cp.GetPhase(time.Now())
	if phase != 3 {
		return false, 0, fmt.Errorf("only phase 3 can get results, current phase: %v", phase)
	}

	conn := cp.pool.Get()
	defer conn.Close()

	// If results have not been generated, generate them.
	exists, err := redis.Bool(conn.Do("EXISTS", "results"))
	if err != nil {
		return false, 0, err
	}

	if !exists {
		if err = cp.GenerateResults(); err != nil {
			return false, 0, err
		}
	}

	price, err := redis.Int64(conn.Do("HGET", "results", ID))
	switch err {
	case redis.ErrNil:
		return false, 0, nil
	case nil:
		return true, price, nil
	default:
		return false, 0, err
	}
}
