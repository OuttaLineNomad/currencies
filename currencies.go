package currencies

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// Rates struct for json response.
type Rates struct {
	Disclaimer string             `json:"disclaimer"`
	License    string             `json:"license"`
	Timestamp  int                `json:"timestamp"`
	Base       string             `json:"base"`
	Rates      map[string]float32 `json:"rates"`
}

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

// ConvertNow gets latest rate and converts the amount.
func (a *APIClient) ConvertNow(from, to string, amount float32) (float32, error) {
	rate, err := a.GetLatestRates(to, from)
	if err != nil {
		return 0, err
	}
	return rate.Rates[to], nil
}

// GetLatestRates gets latest rates for symbols in list, or if list nil sends all rates.
func (a *APIClient) GetLatestRates(base string, symboles ...string) (*Rates, error) {
	return a.rates("latest.json", base, symboles)
}

func (a *APIClient) rates(endpoint, base string, symboles []string) (*Rates, error) {
	url := a.baseURL + "/" + endpoint

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, &Error{Message: "http.NewRequest", Description: err.Error()}
	}
	q := req.URL.Query()
	q.Add("app_id", a.appID)

	if base == "" {
		return a.callAPI(req)
	}

	return nil, nil
}

// callAPI calles api with request given and reters rates.
func (a *APIClient) callAPI(req *http.Request) (*Rates, error) {

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

	rates := &Rates{}

	err = json.Unmarshal(b, rates)
	if err != nil {
		return nil, &Error{Message: "json.Unmarshal() stat 200", Description: err.Error()}
	}

	return rates, nil
}
