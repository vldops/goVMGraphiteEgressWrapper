package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var validateRgxOne = regexp.MustCompile(`graphite{target=.*"(.*)"}`)
var validateRgxTwo = regexp.MustCompile(`.*[\[\]\{\}\*].*`)

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

var victoriaAnswerResult victoriaAnswer

func main() {
	logger.Info("Starting app.")
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/query", goGet).Methods("GET")
	r.HandleFunc("/api/v1/query_range", goGet).Methods("GET")
	log.Fatal(http.ListenAndServe(config.AppPort, r))

}

func goGet(w http.ResponseWriter, r *http.Request) {
	_, ok := r.URL.Query()["query"]
	if ok {

		query := generateQueryStruct(r.URL.Query()["query"][0])

		if query.Error != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Error("", zap.Error(query.Error))
			return
		}

		query.validateRgx(0)

		if query.Level[0].BoolCheck {

			query.checkRegExpValue(0)

			if query.Level[0].Error != nil {
				w.WriteHeader(http.StatusBadRequest)
				logger.Error("Can't Grab RE Value", zap.Error(query.Level[0].Error))
				return
			}

			query.validateRgx(1)
			if query.Level[1].BoolCheck {

				query.checkRegExpValue(1)

				if query.Level[1].Error != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Error("Can't Grab RE Value", zap.Error(query.Level[1].Error))
					return
				}
				query.lexThis(1)
				query.concatThisURL(r.URL.String(), 1)

				if query.Level[1].Error != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Error("Can't Parse Request Query with net/url", zap.Error(query.Level[1].Error))
					return
				}

				bytesResult, statusCode, err := doRequest(query.Level[1].URLStringResult)
				if err != nil {
					logger.Error("Can't Do Request to victoriaMatrics", zap.Error(err))
					w.WriteHeader(statusCode)
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
				logger.Debug("Debug jsonResp", zap.Any("jsonResp", victoriaAnswerResult))
				json.NewEncoder(w).Encode(victoriaAnswerResult)
			} else {
				query.lexThis(0)
				query.concatThisURL(r.URL.String(), 0)

				if query.Level[0].Error != nil {
					w.WriteHeader(http.StatusBadRequest)
					logger.Error("Can't Parse Request Query with net/url", zap.Error(query.Level[1].Error))
					return
				}

				bytesResult, statusCode, err := doRequest(query.Level[0].URLStringResult)
				if err != nil {
					logger.Error("Can't Do Request to victoriaMatrics", zap.Error(err))
					w.WriteHeader(statusCode)
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
				logger.Debug("Debug jsonResp", zap.Any("jsonResp", victoriaAnswerResult))
				json.NewEncoder(w).Encode(victoriaAnswerResult)
			}

		} else {
			bytesResult, statusCode, err := doRequest(query.victoriaURL + r.URL.String())
			if err != nil {
				logger.Error("Can't Do Request to victoriaMatrics", zap.Error(err))
				w.WriteHeader(statusCode)
				return
			}
			w.WriteHeader(statusCode)
			w.Write(bytesResult)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func doRequest(URL string) ([]byte, int, error) {
	client := http.Client{}
	request, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return []byte{}, http.StatusBadRequest, err
	}
	resp, err := client.Do(request)

	if err != nil {
		return []byte{}, resp.StatusCode, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}
