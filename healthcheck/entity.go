package healthcheck

type Health struct {
	Type   string      `json:"type"`
	Status interface{} `json:"status"`
	Err    string      `json:"error,omitempty"`
}

func NewHealth(serviceType string, status interface{}, err error) Health {
	var errMsg string
	if err != nil {
		status = "UNAVAILABLE"
		errMsg = err.Error()
	}
	return Health{
		Type:   serviceType,
		Status: status,
		Err:    errMsg,
	}
}
