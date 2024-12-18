package entities

type Forecast struct {
    Month  int     // Month in the forecasted year
    Value  float64 `json:"Forecasted Demand"` // Forecasted value for the month
}

type EOQ struct {
	EOQ         float64
	HoldingCostRate float64
}


type Cost struct {
	EOQ float64 `json:"For EOQ Total Cost"`
	ShortageAmount float64 `json:"For Shortage Amount Total Cost"`
}
