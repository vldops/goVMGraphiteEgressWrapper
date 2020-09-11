package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var validQuery = regexp.MustCompile(`"(.*)"`)

type victoriaAnswer struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric struct {
				Name   string `json:"__name__"`
				Metric string `json:"metric"`
			} `json:"metric"`
			Values [][]interface{} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type answerStruct struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		Result     []string `json:"result"`
	} `json:"data"`
}

type parserString struct {
	name              string
	revorkQueryString int
	QueryString       string
}

var skippingResult answerStruct
var victoriaAnswerResult victoriaAnswer
var emptySlice = []string{}

func main() {
	logger.Info("Starting app.")
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/query_range", goGet).Methods("GET")
	log.Fatal(http.ListenAndServe(config.AppPort, r))

}

func goGet(w http.ResponseWriter, r *http.Request) {
	logger.Info("Init Requests", zap.String("URL", r.URL.String()))
	_, ok := r.URL.Query()["query"]
	if ok {
		if validQuery.MatchString(r.URL.Query()["query"][0]) {

			modifiedURL, err := url.Parse(r.URL.String())

			if err != nil {
				logger.Error("Can't parse query field", zap.String("QUERY", r.URL.Query()["query"][0]), zap.Error(err))
				w.WriteHeader(http.StatusRequestHeaderFieldsTooLarge)
				return
			}

			modifiedURLValues, err := url.ParseQuery(modifiedURL.RawQuery)

			if err != nil {
				logger.Error("Can't parse query field with url.ParseQuery(modifiedURL.RawQuery)", zap.String("QUERY", modifiedURL.RawQuery), zap.Error(err))
				w.WriteHeader(http.StatusRequestHeaderFieldsTooLarge)
				return
			}

			regExpResult := validQuery.FindAllStringSubmatch(r.URL.Query()["query"][0], 1)[0][1]
			if regExpResult == "" {
				logger.Error("Can't RegExp query field with FindAllStringSubmatch", zap.Strings("RegExpResult", validQuery.FindAllString(r.URL.Query()["query"][0], 1)))
				w.WriteHeader(http.StatusRequestHeaderFieldsTooLarge)
				return
			}
			parserStringResult := parseMePlease(regExpResult)
			modifiedURLValues.Set("query", parserStringResult.QueryString)
			modifiedURL.RawQuery = modifiedURLValues.Encode()
			victoriaMetricsFullRequest := victoriaMetricsAddress + modifiedURL.String()
			bytesResult, err := doRequest(victoriaMetricsFullRequest)
			if err != nil {
				logger.Error("Can't Do Request to victoriaMatrics", zap.Error(err))
				w.WriteHeader(http.StatusRequestHeaderFieldsTooLarge)
				return
			}

			err = json.Unmarshal(bytesResult, &victoriaAnswerResult)
			if err != nil {
				logger.Error("Can't Unmarshal", zap.String("Unmarhal string", string(bytesResult)))
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			for i := range victoriaAnswerResult.Data.Result {
				victoriaAnswerResult.Data.Result[i].Metric.Metric = victoriaAnswerResult.Data.Result[i].Metric.Name
				victoriaAnswerResult.Data.Result[i].Metric.Name = config.MetricName
			}
			w.WriteHeader(http.StatusOK)
			logger.Debug("asd", zap.Any("jsonResp", victoriaAnswerResult))
			json.NewEncoder(w).Encode(victoriaAnswerResult)
		} else {
			skippingResult.Status = "success"
			skippingResult.Data.ResultType = "matrix"
			skippingResult.Data.Result = emptySlice
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(skippingResult)
		}
	}
}

func doRequest(URL string) ([]byte, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return []byte{}, err
	}
	resp, err := client.Do(request)

	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return body, nil
}

func parseMePlease(myString string) parserString {

	parserStringResult := parserString{
		name: myString,
	}

	myRuneSlice := []rune(myString)
	for _, v := range myRuneSlice {
		switch {
		case v == '{':
			parserStringResult.revorkQueryString++
		case v == '}':
			parserStringResult.revorkQueryString++
		case v == '(':
			parserStringResult.revorkQueryString++
		case v == ')':
			parserStringResult.revorkQueryString++
		case v == '|':
			parserStringResult.revorkQueryString++
		case v == ',':
			parserStringResult.revorkQueryString++
		case v == '*':
			parserStringResult.revorkQueryString++
		default:
			continue
		}
	}

	if parserStringResult.revorkQueryString >= 1 {
		parserStringResult.QueryString = fmt.Sprintf("{__name__=~\"%v\"}", myString)
	} else {
		parserStringResult.QueryString = myString
	}
	return parserStringResult
}
