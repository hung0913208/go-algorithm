package oanda

import (
    "encoding/json"
	"io/ioutil"    
    "net/http"
    "time"
	"fmt"
)

type Oanda interface {
    GetCandles(
        instrument  string,
        count       int,
        granularity string,
        from, to    int64,
    ) (*InstrumentHistory, error)
}

type oandaImpl struct {
    hostname       string
	port           int
	ssl            bool
	token          string
	accountID      string
	DatetimeFormat string
	header         *oandaHeaderImpl
    timeout        time.Duration
}

type oandaHeaderImpl struct {
    contentType    string
    agent          string
    datetimeFormat string
    auth           string
}

const OANDA_AGENT string = "v20-golang/0.0.1"

func NewOanda(
    accountID string, 
    token string, 
    live bool,
    timeout time.Duration,
) Oanda { 
    hostname := ""

	// should we use the live API?
	if live {
		hostname = "https://api-fxtrade.oanda.com/v3"
	} else {
		hostname = "https://api-fxpractice.oanda.com/v3"
	}

	// Create headers for oanda to be used in requests
	header := &oandaHeaderImpl{
		contentType:    "application/json",
		agent:          OANDA_AGENT,
		datetimeFormat: "RFC3339",
		auth:           fmt.Sprintf("Bearer %s", token),
	}

	// Create the connection object
    if timeout == 0 {
	    return &oandaImpl{
	    	hostname:  hostname,
	    	port:      443,
	    	ssl:       true,
	    	token:     token,
	    	header:    header,
	    	accountID: accountID,
            timeout:   time.Second * time.Duration(5),
	    }
    } else {
	    return &oandaImpl{
	    	hostname:  hostname,
	    	port:      443,
	    	ssl:       true,
	    	token:     token,
	    	header:    header,
	    	accountID: accountID,
            timeout:   timeout,
	    }
    }
}

func (self *oandaImpl) GetCandles(
    instrument  string,
    count       int,
    granularity string,
    from, to    int64,
) (*InstrumentHistory, error) { 
    var candles InstrumentHistory

    query := fmt.Sprintf(
        "/instruments/%s/candles?count=%d&granularity=%s",
        instrument,
        count,
        granularity,
    )

    if from < to && from > 0 {
        query = fmt.Sprintf(
            "/instruments/%s/candles?count=%d&granularity=%s&from=%d&to=%d",
            instrument,
            count,
            granularity,
            from, to,
        )
    }

	resp, err := self.get(query)
    if err != nil {
        return nil, err
    }
    
	body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

	err = json.Unmarshal(body, &candles)
    if err != nil {
        return nil, err
    }

    return &candles, nil
}


func (self *oandaImpl) get(endpoint string) (*http.Response, error) {
	client := http.Client{
		Timeout: self.timeout,
	}

	// New request object
	req, err := http.NewRequest(
        http.MethodGet, 
        fmt.Sprintf("%s/%s", self.hostname, endpoint), 
        nil,
    )
    if err != nil {
        return nil, err
    }

	req.Header.Set("User-Agent", self.header.agent)
	req.Header.Set("Content-Type", self.header.contentType)
	req.Header.Set("Authorization", self.header.auth)

	return client.Do(req)
}
