package chepai

type Reply struct {
	Success bool   `json:"success"`
	ErrMsg  string `json:"err"`
	ID      string `json:"id"`
}

type UnixNanoTimeInfoReply struct {
	Reply
	BeginTime       int64 `json:"begin_time"`
	PhaseOneEndTime int64 `json:"phase_one_end_time"`
	PhaseTwoEndTime int64 `json:"phase_two_end_time"`
}

type UnixTimeInfoReply struct {
	Reply
	BeginTime       int64 `json:"begin_time"`
	PhaseOneEndTime int64 `json:"phase_one_end_time"`
	PhaseTwoEndTime int64 `json:"phase_two_end_time"`
}

type StartPriceReply struct {
	Reply
	StartPrice int64 `json:"start_price"`
}

type LicensePlateNumReply struct {
	Reply
	LicensePlateNum int64 `json:"license_plate_num"`
}

type LowestPriceReply struct {
	Reply
	LowestPrice int64 `json:"lowest_price"`
}

type BidderNumReply struct {
	Reply
	BidderNum int64 `json:"bidder_num"`
}

type BidReply struct {
	Reply
	Phase int   `json:"phase"`
	Price int64 `json:"price"`
}

type BidRecordsReply struct {
	Reply
	Records []*BidRecord `json:"bid_records"`
}

type ResultsReply struct {
	Reply
	Results map[string]string `json:"results"`
}

type ResultReply struct {
	Reply
	Done  bool  `json:"done"`
	Price int64 `json:"price"`
}
