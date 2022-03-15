package conversion

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type ConversionService struct {
	client *http.Client
}

func NewService() *ConversionService {
	cli := http.Client{}
	s := &ConversionService{client: &cli}
	return s
}

func (cs *ConversionService) GetRate() (float64, error) {
	resp, err := cs.getRate()
	if err != nil {
		return 0, err
	}

	rate, err := parseResponse(resp)
	if err != nil {
		return 0, err
	}
	return rate, nil
}

func (cs *ConversionService) getRate() (map[string]interface{}, error) {
	request, err := http.NewRequest("GET", "https://www.cbr-xml-daily.ru/daily_json.js", nil)
	if err != nil {
		return nil, err
	}

	resp, err := cs.client.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("не удалось получить курс")
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result, nil
}

func parseResponse(resp map[string]interface{}) (float64, error) {
	if rates := resp["Valute"]; rates != nil {
		t := reflect.ValueOf(rates)
		iter := t.MapRange()
		for iter.Next() {
			if key := iter.Key(); key.String() == "USD" {
				fmt.Println(iter.Value())
				mapUSD := iter.Value().Elem().MapRange()
				for mapUSD.Next() {
					if key := mapUSD.Key(); key.String() == "Value" {
						rate := mapUSD.Value().Elem().Float()
						fmt.Println(rate)
						return rate, nil
					}
				}
			}
		}
	}
	return 0, errors.New("не удалось получить курс")
}
