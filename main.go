package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hashicorp/go-hclog"
	protos "github.com/rassulmagauin/workplace/grpcpractice/protos/currency/protos/currency"

	"github.com/rassulmagauin/workplace/data"
	"github.com/rassulmagauin/workplace/handlers"
	"github.com/rassulmagauin/workplace/middlewares"
	"google.golang.org/grpc"
)

func main() {
	l := hclog.Default()
	conn, err := grpc.Dial("localhost:9092", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	cc := protos.NewCurrencyClient(conn)

	db := data.NewProductsDB(l, cc)

	ph := handlers.NewProducts(l, db)
	fh := handlers.NewFiles(l)
	zm := middlewares.NewZipMiddleware()
	jm := middlewares.NewJwtMiddleware()
	sm := mux.NewRouter()

	getRouter := sm.Methods("GET").Subrouter()
	getRouter.HandleFunc("/", ph.GetProducts).Queries("currency", "{[A-Z]{3}}")
	getRouter.HandleFunc("/", ph.GetProducts)
	getRouter.HandleFunc("/{id:[0-9]+}", ph.GetProduct).Queries("currency", "{[A-Z]{3}}")
	getRouter.HandleFunc("/{id:[0-9]+}", ph.GetProduct)

	putRouter := sm.Methods(http.MethodPut).Subrouter()
	putRouter.HandleFunc("/{id:[0-9]+}", ph.UpdateProduct)
	putRouter.Use(ph.Middleware)

	postRouter := sm.Methods("POST").Subrouter()
	postRouter.HandleFunc("/", ph.AddProduct)
	postRouter.Use(ph.Middleware)

	fp := sm.Methods("POST").Subrouter()
	fp.HandleFunc("/upload/{id:[0-9]+}", fh.UploadFile)

	fg := sm.Methods("GET").Subrouter()
	fg.HandleFunc("/images/{id:[0-9]+}", fh.GetFile)
	fg.Use(jm.CheckJWT)
	fg.Use(zm.GzipMiddleware)

	tg := sm.Methods("GET").Subrouter()
	tg.HandleFunc("/token", jm.GetJWT)

	//CORS
	ch := gohandlers.CORS(gohandlers.AllowedOrigins([]string{"http://localhost:3000"}))

	s := http.Server{
		Addr:         ":8000",
		Handler:      ch(sm),
		ErrorLog:     l.StandardLogger(&hclog.StandardLoggerOptions{}),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	go func() {
		l.Info("Starting server on prot 8000")
		err := s.ListenAndServe()
		if err != nil {
			l.Error("Error starting server", "error", err)
			os.Exit(1)
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, os.Kill)

	sig := <-sigChan
	l.Info("Received terminate, graceful shutdown", sig)

	tc, _ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)

}
