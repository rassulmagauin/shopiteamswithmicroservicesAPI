package server

import (
	"context"
	"io"
	"time"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/rassulmagauin/grpcpractice/data"
	protos "github.com/rassulmagauin/grpcpractice/protos/currency/protos/currency"
)

type Currency struct {
	protos.UnimplementedCurrencyServer
	log           hclog.Logger
	rates         *data.ExchangeRates
	subscriptions map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest
}

func NewCurrency(l hclog.Logger, r *data.ExchangeRates) *Currency {
	с := &Currency{log: l, rates: r, subscriptions: make(map[protos.Currency_SubscribeRatesServer][]*protos.RateRequest)}
	go с.handleUpdates()
	return с
}

func (c *Currency) handleUpdates() {
	ru := c.rates.MonitorRates(time.Second * 5)
	for range ru {
		c.log.Info("Got updated rates")
		for k, v := range c.subscriptions {
			for _, rr := range v {
				r, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
				if err != nil {
					c.log.Error("Unable to get updated rate", "base", rr.GetBase(), "destination", rr.GetDestination())
				}
				err = k.Send(&protos.RateResponse{Base: rr.GetBase(), Destination: rr.GetDestination(), Rate: r})
				if err != nil {
					c.log.Error("Unable to send updated rate", "base", rr.GetBase(), "destination", rr.GetDestination())
				}
			}

		}
	}
}

func (c *Currency) GetRate(ctx context.Context, rr *protos.RateRequest) (*protos.RateResponse, error) {
	c.log.Info("Handle GetRate", "base", rr.GetBase(), "destination", rr.GetDestination())
	rate, err := c.rates.GetRate(rr.GetBase().String(), rr.GetDestination().String())
	if err != nil {
		c.log.Error("Unable to get rate", "error", err)
		return nil, err
	}
	return &protos.RateResponse{Base: rr.Base, Destination: rr.Destination, Rate: rate}, nil
}

func (c *Currency) SubscribeRates(srs protos.Currency_SubscribeRatesServer) error {
	for {
		rr, err := srs.Recv()
		if err == io.EOF {
			c.log.Info("Client has closed connection")
			return err
		}
		if err != nil {
			c.log.Error("Unable to read from client", "error", err)
			return err
		}
		c.log.Info("Handle client request", "request_base", rr.GetBase(), "request_destination", rr.GetDestination())
		rrs, ok := c.subscriptions[srs]
		if !ok {
			rrs = []*protos.RateRequest{}
		}
		rrs = append(rrs, rr)
		c.subscriptions[srs] = rrs
	}
	return nil
}
