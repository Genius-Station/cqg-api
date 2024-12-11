package cqgapi

import (
	"math"
)

func round(amount float64, tick float64, precision float64) float64 {
	if tick > 0 {
		return math.Round(amount/tick) * tick
	}
	return roundToPrecision(amount, precision)
}


func roundToPrecision(amount float64, precision float64) float64 {
	return math.Round(amount*math.Pow(10, precision)) / math.Pow(10, precision)
}

func getRiskReward(alert Alert, tick float64, precision float64) float64 {
	start := round(alert.Start, tick, precision)
	stop := round(alert.Stop, tick, precision)
	target := round(alert.Target, tick, precision)

	if start == stop {
		return 0
	}
	if start > target {
		return (start - target) / (stop - start)
	} else {
		return (target - start) / (start - stop)
	}
}

func getMarginPercentage(riskReward float64) float64 {
	if riskReward < 1.5 {
		return 25
	}
	if riskReward < 2 {
		return 50
	}
	if riskReward < 2.5 {
		return 75
	}
	return 100
}


func GetQuantity(availableMargin float64, alert Alert , maxLots *int64) int {
	tick := alert.tick // todo 
	precision := alert.PricePrecision
	riskReward := getRiskReward(alert, tick, precision)
	marginPercentage := getMarginPercentage(riskReward)
	leverageSpot := alert.leverage // todo leverage du spot
	if leverageSpot == 0 {
		leverageSpot = 1
	}

	
	maxQuantity := int(math.Floor(availableMargin / (alert.Start * (1 + 0.05))))

	
	maxAmountRR := availableMargin * marginPercentage / 100 / leverageSpot

	
	quantity := int(math.Floor(maxAmountRR / alert.Start))

	
	if quantity == 0 {
		return quantity
	}

	// todo lotsAvg
	if alert.lotsAvg > 0 {
		if quantity > alert.lotsAvg {
			quantity = alert.lotsAvg
		}
	}

	
	if maxLots != nil && *maxLots > 0 {
		if quantity > int(*maxLots) {
			quantity = int(*maxLots)
		}
	}

	
	if quantity > maxQuantity {
		return maxQuantity
	}
	return quantity
}
