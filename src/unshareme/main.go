package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
)

var hashKey = securecookie.GenerateRandomKey(32)
var blockKey = securecookie.GenerateRandomKey(32)
var name = "name"
var sc = securecookie.New(hashKey, blockKey)
var router = mux.NewRouter()

func encode(msg string, r *http.Request) (string, error) {
	msg += "|" + remoteIP(r)
	enc, err := sc.Encode(name, msg)
	if err != nil {
		return "", err
	}

	b64enc := base64.URLEncoding.EncodeToString([]byte(enc))

	return b64enc, nil
}

func decode(enc string, r *http.Request) (string, error) {
	b64enc, err := base64.URLEncoding.DecodeString(enc)
	if err != nil {
		return "", err
	}

	var msg string
	err = sc.Decode(name, string(b64enc), &msg)
	if err != nil {
		return "", err
	}

	tokens := strings.Split(msg, "|")
	if tokens[1] != remoteIP(r) {
		return "", errors.New("Invalid IP")
	}

	return tokens[0], nil
}

func remoteIP(r *http.Request) string {
	return strings.Split(r.RemoteAddr, "]")[0][1:]
}

func EncodeHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")

	enc, err := encode(msg, r)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	link, _ := router.Get("Decode").URL("enc", enc)
	fmt.Fprintf(w, "<a href=\"%s\">Link</a>", link.String())
//         fmt.Fprint(w, link.String())
}

func DecodeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dec, err := decode(vars["enc"], r)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, dec, http.StatusFound)
	return
}

func main() {
	router.Handle("/favicon.ico", http.NotFoundHandler())
	router.HandleFunc("/", EncodeHandler).Methods("GET")
	router.HandleFunc("/{enc}", DecodeHandler).Methods("GET").Name("Decode")
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":7001", nil))
}
