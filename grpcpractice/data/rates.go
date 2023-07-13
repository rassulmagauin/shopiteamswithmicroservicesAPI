package data

import (
	"encoding/xml"
	"errors"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	hclog "github.com/hashicorp/go-hclog"
)

type ExchangeRates struct {
	l     hclog.Logger
	rates map[string]float64
}

func NewRates(l hclog.Logger) (*ExchangeRates, error) {
	er := &ExchangeRates{l: l, rates: map[string]float64{}}
	err := er.GetRates()
	return er, err
}

func (e *ExchangeRates) GetRate(base, dest string) (float64, error) {
	ErrRateNotFound := errors.New("rate not found")
	br, ok := e.rates[base]
	if !ok {

		return 0, ErrRateNotFound
	}
	dr, ok := e.rates[dest]
	if !ok {
		return 0, ErrRateNotFound
	}
	return dr / br, nil
}
func (e *ExchangeRates) GetRates() error {
	rates, err := http.Get("https://www.ecb.europa.eu/stats/eurofxref/eurofxref-daily.xml")
	if err != nil {
		return err
	}
	defer rates.Body.Close()

	md := &Cubes{}
	xml.NewDecoder(rates.Body).Decode(md)
	for _, c := range md.Cubes {
		rate, err := strconv.ParseFloat(c.Rate, 64)
		if err != nil {
			return err
		}
		e.rates[c.Currency] = rate
	}
	e.rates["EUR"] = 1
	return nil
}

func (e *ExchangeRates) MonitorRates(interval time.Duration) chan struct{} {
	ret := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-ticker.C:
				// just add a random difference to the rate and return it
				// this simulates the fluctuations in currency rates
				for k, v := range e.rates {
					// change can be 10% of original value
					change := (rand.Float64() / 10)
					// is this a postive or negative change
					direction := rand.Intn(1)

					if direction == 0 {
						// new value with be min 90% of old
						change = 1 - change
					} else {
						// new value will be 110% of old
						change = 1 + change
					}

					// modify the rate
					e.rates[k] = v * change
				}

				// notify updates, this will block unless there is a listener on the other end
				ret <- struct{}{}
			}
		}
	}()

	return ret
}

type Cubes struct {
	Cubes []Cube `xml:"Cube>Cube>Cube"`
}

type Cube struct {
	Currency string `xml:"currency,attr"`
	Rate     string `xml:"rate,attr"`
}
