package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func stringtoUint(s string) (uint, error) {
	i, err := strconv.Atoi(s)
	return uint(i), err
}

func WriteError(w http.ResponseWriter, err error, errorCode int) {
	log.Println(err)
	errorMap := make(map[string]string)
	errorMap["Error"] = fmt.Sprintf("%s", err)
	jsonOut, _ := json.Marshal(errorMap)
	w.WriteHeader(errorCode)
	w.Write(jsonOut)
}

func WriteErrorString(w http.ResponseWriter, errString string, errorCode int) {
	WriteError(w, errors.New(errString), errorCode)
}

func WriteJson(w http.ResponseWriter, input interface{}) {
	jsonOut, err := json.Marshal(input)
	if err != nil {
		WriteError(w, err, 500)
		return
	}
	w.Write(jsonOut)
}

func WriteOutputError(w http.ResponseWriter, input interface{}, err error) {
	if err != nil {
		WriteError(w, err, 500)
		return
	}
	WriteJson(w, input)
}

type pair struct {
	key   string
	value interface{}
}

//func genMap(key ...string, value ...interface{}) map[string]interface{} {
//func genMap(pairs ...pair) map[string]interface{} {
//	m := make(map[string]interface{})
//	for index := 0; index < len(pairs); index++ {
//wtf do i do how do i fix mark haaaaalp
//	}
//	for index := 0; index < len(key); index++ {
//		m[key[index]] = value[index]
//	}
//	return m
func genMap(key string, value interface{}) map[string]interface{} {
	return map[string]interface{}{key: value}
}
