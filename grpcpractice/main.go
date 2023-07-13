package main

import (
	"net"

	hclog "github.com/hashicorp/go-hclog"
	"github.com/rassulmagauin/grpcpractice/data"
	protos "github.com/rassulmagauin/grpcpractice/protos/currency/protos/currency"
	"github.com/rassulmagauin/grpcpractice/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	log := hclog.Default()
	gs := grpc.NewServer()
	er, err := data.NewRates(log)
	if err != nil {
		log.Error("Unable to generate rates", "error", err)
	}
	cs := server.NewCurrency(log, er)
	protos.RegisterCurrencyServer(gs, cs)

	reflection.Register(gs)

	l, err := net.Listen("tcp", ":9092")
	if err != nil {
		log.Error("Unable to listen", "error", err)
	}

	gs.Serve(l)

}
