package stat

type Stat struct {
	ServiceType string      `json:"type"`
	Stat        interface{} `json:"stat"`
	Err         string      `json:"error,omitempty"`
}

func NewStat(serviceType string, stat interface{}, err error) Stat {
	var errMsg string
	if err != nil {
		stat = "UNAVAILABLE"
		errMsg = err.Error()
	}
	return Stat{
		ServiceType: serviceType,
		Stat:        stat,
		Err:         errMsg,
	}
}
