package currencies

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"
)

// RatesResponse struct for json response.
type RatesResponse struct {
	Disclaimer string `json:"disclaimer"`
	License    string `json:"license"`
	Timestamp  int    `json:"timestamp"`
	Base       string `json:"base"`
	Rates      Rates  `json:"rates"`
}

// Rates map of symboles and rates from api.
type Rates map[string]float64

// APIClient base for currencies package.
type APIClient struct {
	baseURL string
	appID   string
}

// Error struct of external api error.
type Error struct {
	IsError     bool   `json:"error"`
	Status      int    `json:"status"`
	Message     string `json:"message"`
	Description string `json:"description"`
}

func (er *Error) Error() string {
	return "currencies: " + er.Message + ": " + er.Description
}

// New creates instance of currencies.
func New(appID string) *APIClient {
	return &APIClient{"https://openexchangerates.org/api/", appID}
}

// ConvertNow gets latest rate and converts the amount in base currency.
func (a *APIClient) ConvertNow(from, to string, amt float64) (float64, error) {
	rates, err := a.GetLatestRates(from, to)
	if err != nil {
		return 0, err
	}

	rate := rates[to]
	return math.Round((amt*rate)*100) / 100, nil
}

// GetLatestRates gets latest rates for symbols in list, or if list nil sends all rates.
func (a *APIClient) GetLatestRates(base string, symboles ...string) (Rates, error) {
	rates, err := a.rates("latest.json", base, symboles)
	if err != nil {
		return nil, err
	}
	return rates.Rates, nil
}

func (a *APIClient) rates(endpoint, base string, symboles []string) (*RatesResponse, error) {
	url := a.baseURL + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &Error{Message: "http.NewRequest", Description: err.Error()}
	}
	q := req.URL.Query()
	q.Add("app_id", a.appID)

	if base == "" {
		req.URL.RawQuery = q.Encode()
		return a.callAPI(req)
	}

	if symboles == nil {
		q.Add("base", base)
		req.URL.RawQuery = q.Encode()
		return a.callAPI(req)
	}

	symbolesStr := strings.Join(symboles, ",")
	q.Add("symboles", symbolesStr)
	req.URL.RawQuery = q.Encode()

	return a.callAPI(req)
}

// callAPI calles api with request given and reters rates.
func (a *APIClient) callAPI(req *http.Request) (*RatesResponse, error) {

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, &Error{Message: "client.Do()", Description: err.Error()}
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		err = &Error{}
		err = json.Unmarshal(b, err)
		if err != nil {
			return nil, &Error{Message: "json.Unmarshal() stat !200", Description: err.Error()}
		}
		return nil, err
	}

	rates := &RatesResponse{}

	err = json.Unmarshal(b, rates)
	if err != nil {
		return nil, &Error{Message: "json.Unmarshal() stat 200", Description: err.Error()}
	}
	return rates, nil
}
