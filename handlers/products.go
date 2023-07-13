// Package classification Producs API.
//
// the purpose of this application is to provide an application
// that is using plain go code to define an API
//
// This should demonstrate all the possible comment annotations
// that are available to turn go code into a fully compliant swagger 2.0 spec
//
// Terms Of Service:
//
// there are no TOS at this moment, use at your own risk we take no responsibility
//
//	Schemes: http
//	Host: localhost
//	BasePath: /
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
// swagger:meta
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	hclog "github.com/hashicorp/go-hclog"
	"github.com/rassulmagauin/workplace/data"
)

type Products struct {
	l         hclog.Logger
	productDB *data.ProducstsDB
}

func NewProducts(l hclog.Logger, productDB *data.ProducstsDB) *Products {
	return &Products{l, productDB}
}

// swagger:route GET /products products listProducts
// Returns list of products
// responses:
//   200: productsResponse

func (p *Products) GetProducts(rw http.ResponseWriter, r *http.Request) {
	p.l.Debug("Handle GET Products")
	cur := r.URL.Query().Get("currency")
	lp, err := p.productDB.GetProducts(cur)
	if err != nil {
		http.Error(rw, "Unable to get products", http.StatusInternalServerError)
		return
	}
	rw.Header().Set("Content-Type", "application/json")

	err = lp.ToJSON(rw)
	if err != nil {
		p.l.Error("Unable to marshal json", "error", err)
		http.Error(rw, "Unable to Marshal JSON", http.StatusInternalServerError)
	}
}

func (p *Products) GetProduct(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	cur := r.URL.Query().Get("currency")
	if err != nil {
		http.Error(rw, "Unable to find id", http.StatusBadRequest)
		return
	}
	p.l.Debug("Handle GET Product", "id", id)
	prod, err := p.productDB.GetProduct(id, cur)
	if err != nil {
		http.Error(rw, "Unable to find product", http.StatusNotFound)
		return
	}
	rw.Header().Set("Content-Type", "application/json")
	err = prod.ToJSON(rw)
	if err != nil {
		p.l.Error("Unable to Marhsal JSON", err)
		http.Error(rw, "Unable to Marshal JSON", http.StatusInternalServerError)
	}
}

func (p *Products) AddProduct(rw http.ResponseWriter, r *http.Request) {
	p.l.Debug("Handling POST request")
	prod := r.Context().Value("prod").(data.Product)
	data.AddProduct(&prod)

}

func (p *Products) UpdateProduct(rw http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(rw, "Unable to find id", http.StatusBadRequest)
		return
	}
	prod := r.Context().Value("prod").(data.Product)
	err = data.ReplaceProduct(id, prod)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusNotFound)
	}
	rw.WriteHeader(http.StatusOK)
}

func (p Products) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		prod := data.Product{}
		d, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(rw, "Unable to read Body", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(d, &prod)
		if err != nil {
			http.Error(rw, "Unable to Unmarshal", http.StatusBadRequest)
			return
		}
		err = prod.Validate()
		if err != nil {
			http.Error(
				rw,
				fmt.Sprint("Error in Validation :", err),
				http.StatusBadRequest,
			)
			return
		}
		ctx := context.WithValue(r.Context(), "prod", prod)
		r = r.WithContext(ctx)
		next.ServeHTTP(rw, r)
	})
}

// A list of Products
// swagger:response productsResponse
type productsResponse struct {
	//All products in the system
	//in:body
	Body []data.Product
}
