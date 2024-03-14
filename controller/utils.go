package controller

import (
	"net/http"
	"strings"

	"github.com/bldsoft/gost/log"
	"github.com/bldsoft/gost/repository"
	"github.com/bldsoft/gost/utils"
)

func ResponseError(w http.ResponseWriter, err string, code int) {
	if code >= 500 {
		if errWriter, ok := log.AsResponseWriterLogErr(w); ok {
			errWriter.WriteRequestInfoErr(err)
		}
	}
	http.Error(w, err, code)
}

// GetQueryOption doesn't return error a query param is successfully parsed or not present
func GetQueryOption[T utils.Parsed](r *http.Request, paramName string, defaultValue ...T) (result T, err error) {
	if len(defaultValue) > 0 {
		result = defaultValue[0]
	}

	if strValue := r.URL.Query().Get(paramName); strValue != "" {
		return utils.Parse[T](strValue)
	}
	return result, nil
}

// GetQuerySlice doesn't return error if query param is successfully parsed or not present
func GetQueryOptionSlice[T utils.Parsed](r *http.Request, paramName string) (result []T, err error) {
	if strValues := r.URL.Query().Get(paramName); strValues != "" {
		values := strings.Split(strValues, ",")
		for _, strValue := range values {
			if value, err := utils.Parse[T](strValue); err != nil {
				return nil, err
			} else {
				result = append(result, value)
			}
		}
	}
	return result, nil
}

// ParseQueryOption returns true if query param is successfully parsed or not present
func ParseQueryOption[T utils.Parsed](r *http.Request, w http.ResponseWriter, paramName string, outValue *T) (ok bool) {
	value, err := GetQueryOption(r, paramName, *outValue)
	if err != nil {
		log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err, "param name": paramName}, "failed to parse query param")
		ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return false
	}
	*outValue = value
	return true
}

// ParseQueryOptionSlice returns true if query param is successfully parsed or not present
func ParseQueryOptionSlice[T utils.Parsed](r *http.Request, w http.ResponseWriter, paramName string, outValue *[]T) (ok bool) {
	value, err := GetQueryOptionSlice[T](r, paramName)
	if err != nil {
		log.FromContext(r.Context()).ErrorWithFields(log.Fields{"err": err, "param name": paramName}, "failed to parse query param")
		ResponseError(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return false
	}
	*outValue = value
	return true
}

func QueryOptionFromRequest[F any](r *http.Request, archived bool) (*repository.QueryOptions, error) {
	opt := &repository.QueryOptions{}
	var err error
	opt.Filter, err = utils.FromRequest[F](r)
	if err != nil {
		return nil, err
	}
	opt.Archived = archived

	fields, err := GetQueryOptionSlice[string](r, "fields")
	if err != nil {
		return nil, err
	}
	opt.Fields = fields

	return opt, nil
}
