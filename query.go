package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"

	"go.uber.org/zap"
)

type queryStruct struct {
	QueryName   string
	Error       error
	victoriaURL string
	Level       [2]struct {
		Regexp          *regexp.Regexp
		BoolCheck       bool
		Result          string
		Error           error
		LexResult       string
		URLStringRAW    string
		URLStringResult string
	}
}

func (q *queryStruct) validateRgx(id int) {
	q.Level[id].BoolCheck = q.Level[id].Regexp.MatchString(q.QueryName)
}

func generateQueryStruct(queryName string) queryStruct {
	if queryName == "" {
		query := queryStruct{}
		query.Error = errors.New("Empty Query")
		return query
	}
	query := queryStruct{}
	query.QueryName = queryName
	query.victoriaURL = victoriaMetricsAddress
	query.Level[0].Regexp = validateRgxOne
	query.Level[1].Regexp = validateRgxTwo
	return query
}

func (q *queryStruct) checkRegExpValue(id int) {

	switch id {
	case 0:
		q.Level[id].Result = q.Level[id].Regexp.FindAllStringSubmatch(q.QueryName, 1)[0][id+1]
		logger.Debug("Check RegExp Result Level 0", zap.Any("rgxResult", q.Level[id].Result), zap.String("object", q.QueryName))
		if q.Level[id].Result == "" {
			q.Level[id].Error = errors.New("Can't RegExp query field with FindAllStringSubmatch")
		}
		q.Level[id].Error = nil

	case 1:
		q.Level[id].Result = q.Level[id].Regexp.FindAllStringSubmatch(q.Level[0].Result, 1)[0][id-1]
		logger.Debug("Check RegExp Result Level 1", zap.Any("rgxResult", q.Level[id].Regexp.FindAllStringSubmatch(q.Level[0].Result, 1)[0][0]), zap.String("object", q.Level[0].Result))
		if q.Level[id].Result == "" {
			q.Level[id].Error = errors.New("Can't RegExp query field with FindAllStringSubmatch")
		}
		q.Level[id].Error = nil

	}

}

func (q *queryStruct) lexThis(id int) {
	if q.Level[id].BoolCheck {
		myRuneSlice := []rune(q.Level[id].Result)
		myRuneSliceRework := []rune{}
		for _, v := range myRuneSlice {
			switch {
			case v == '{':
				myRuneSliceRework = append(myRuneSliceRework, '(')
			case v == '}':
				myRuneSliceRework = append(myRuneSliceRework, ')')
			case v == ',':
				myRuneSliceRework = append(myRuneSliceRework, '|')
			case v == '*':
				myRuneSliceRework = append(myRuneSliceRework, '.', '*')
			default:
				myRuneSliceRework = append(myRuneSliceRework, v)
			}
		}

		q.Level[id].LexResult = fmt.Sprintf("{__name__=~\"%v\"}", string(myRuneSliceRework))
	} else {
		q.Level[id].LexResult = fmt.Sprintf("{__name__=\"%v\"}", q.Level[id].Result)
	}
	logger.Debug("lex results", zap.String("lexResult", q.Level[id].LexResult))
}

func (q *queryStruct) concatThisURL(URL string, id int) {
	q.Level[id].URLStringRAW = URL
	modifiedURL, err := url.Parse(URL)

	if err != nil {
		q.Level[id].Error = err
		return
	}

	modifiedURLValues, err := url.ParseQuery(modifiedURL.RawQuery)

	if err != nil {
		q.Level[id].Error = err
		return
	}

	modifiedURLValues.Set("query", q.Level[1].LexResult)
	modifiedURL.RawQuery = modifiedURLValues.Encode()
	q.Level[id].URLStringResult = q.victoriaURL + modifiedURL.String()
}
