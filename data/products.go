package data

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"regexp"

	"time"

	"github.com/go-playground/validator/v10"
	hclog "github.com/hashicorp/go-hclog"
	protos "github.com/rassulmagauin/workplace/grpcpractice/protos/currency/protos/currency"
)

// swagger:model
type Product struct {
	// the id for this product
	// required: true
	// min: 1
	ID int `json:"id"`

	// the name of this product
	// required: true
	// min length: 1
	Name string `json:"name" validate:"required"`

	// the description of this product
	// required: false
	Description string `json:"description"`

	// the price of this product
	// required: true
	// min: 0.01
	Price float64 `json:"price"`

	// the SKU of this product
	// required: true
	// pattern: ^[a-zA-Z0-9_-]*$
	SKU string `json:"sku" validate:"required,sku"`

	// the creation date of this product
	// required: false
	CreatedOn string `json:"-"`

	// the last updated date of this product
	// required: false
	UpdatedOn string `json:"-"`

	// the deleted date of this product
	// required: false
	DeletedOn string `json:"-"`
}

type ProducstsDB struct {
	l      hclog.Logger
	cc     protos.CurrencyClient
	rates  map[string]float64
	client protos.Currency_SubscribeRatesClient
}

func NewProductsDB(l hclog.Logger, cc protos.CurrencyClient) *ProducstsDB {
	pb := &ProducstsDB{l, cc, make(map[string]float64), nil}
	go pb.handleUpdates()
	return pb
}
func (p *ProducstsDB) handleUpdates() {
	sub, err := p.cc.SubscribeRates(context.Background())
	if err != nil {
		p.l.Error("Unable to subscribe to rates", "error", err)
		return
	}
	p.client = sub
	for {
		rr, err := sub.Recv()
		p.l.Debug("Received updated rate from server", "dest", rr.GetDestination(), "rate", rr.GetRate())
		if err != nil {
			p.l.Error("Unable to receive rates", "error", err)
			continue
		}
		p.rates[rr.GetDestination().String()] = rr.GetRate()
	}

}
func (p *ProducstsDB) GetProducts(currency string) (Products, error) {
	if string(currency) == "" {
		return productList, nil
	}
	rate, err := p.GetRate(currency)
	if err != nil {
		p.l.Error("Unable to get rate", "error", err)
		return nil, err
	}

	pl := make([]*Product, len(productList))
	for i, product := range productList {
		newProduct := *product   // Create a new instance
		newProduct.Price *= rate // Modify the price of the new product
		pl[i] = &newProduct      // Add the new product to the new slice
	}
	return pl, nil
}

func validateSKU(fl validator.FieldLevel) bool {
	re := regexp.MustCompile(`[a-z]+-[a-z]+-[a-z]+`)
	matches := re.FindAllString(fl.Field().String(), -1)

	if len(matches) != 1 {
		return false
	}
	return true
}
func getNextID() int {
	return (productList[len(productList)-1].ID + 1)
}
func (p *Product) Validate() error {
	validate := validator.New()
	validate.RegisterValidation("sku", validateSKU)
	return validate.Struct(p)
}

func AddProduct(p *Product) {
	p.ID = getNextID()
	productList = append(productList, p)
}

type Products []*Product

func (p *ProducstsDB) GetProduct(id int, currency string) (*Product, error) {
	if string(currency) == "" {
		i := findID(id)
		if i < 0 {
			return nil, errors.New("No that ID!")
		}
		return productList[i], nil
	}
	rate, err := p.GetRate(currency)
	if err != nil {
		p.l.Error("Unable to get rate", "error", err)
		return nil, err
	}
	i := findID(id)
	if i < 0 {
		return nil, errors.New("No that ID!")
	}
	pr := *productList[i]
	pr.Price = pr.Price * rate
	return &pr, nil

}

func (p *Products) ToJSON(w io.Writer) error {

	e := json.NewEncoder(w)
	return e.Encode(p)
}

func (p *Product) FromJSON(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(p)
}
func (p *Product) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(p)
}

func ReplaceProduct(id int, prod Product) error {
	i := findID(id)
	if i < 0 {
		return errors.New("No that ID!")
	}
	*productList[i] = prod
	return nil
}

func findID(id int) int {
	for i, x := range productList {
		if x.ID == id {
			return i
		}
	}
	return -1
}

var productList = []*Product{
	&Product{
		ID:          1,
		Name:        "Latte",
		Description: "Frothy milky coffee",
		Price:       2.45,
		SKU:         "abc323",
		CreatedOn:   time.Now().UTC().String(),
		UpdatedOn:   time.Now().UTC().String(),
	},
	&Product{
		ID:          2,
		Name:        "Espresso",
		Description: "Short and strong coffee without milk",
		Price:       1.99,
		SKU:         "fjd34",
		CreatedOn:   time.Now().UTC().String(),
		UpdatedOn:   time.Now().UTC().String(),
	},
}

func (p *ProducstsDB) GetRate(currency string) (float64, error) {
	if _, ok := p.rates[currency]; ok {
		return p.rates[currency], nil
	}

	r := &protos.RateRequest{
		Base:        protos.Currencies(protos.Currencies_value["USD"]),
		Destination: protos.Currencies(protos.Currencies_value[currency]),
	}
	//get inital rate
	resp, err := p.cc.GetRate(context.Background(), r)
	p.rates[currency] = resp.Rate

	//subscribe to updates
	p.client.Send(r)

	if err != nil {
		return 0, err
	}
	return resp.Rate, nil
}
