package main

import (
	"encoding/base64"
	"fmt"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Random stuff for encoding
var hashKey = securecookie.GenerateRandomKey(32)
var blockKey = securecookie.GenerateRandomKey(32)
var encodeName = "encodeName"
var sc = securecookie.New(hashKey, blockKey)

// Router for handlers
var router = mux.NewRouter()

// Store URI and IP together
type PersonalURL struct {
	URI string
	IP string
}

func encode(msg PersonalURL) (string, error) {
	enc, err := sc.Encode(encodeName, msg)
	if err != nil {
		return "", err
	}

	b64enc := base64.URLEncoding.EncodeToString([]byte(enc))

	return b64enc, nil
}

func decode(enc string) (msg PersonalURL, err error) {
	b64enc, err := base64.URLEncoding.DecodeString(enc)
	if err != nil {
		return
	}

	err = sc.Decode(encodeName, string(b64enc), &msg)
	if err != nil {
		return
	}

	return
}

// Only works for IPv4, like 127.0.0.1:12345, not IPv6 like [::1]:12345
func remoteIP(r *http.Request) string {
	// Get it from headers, as set by nginx
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		// Strips port number
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	log.Print("IP:", ip)
	return ip
}

func EncodeHandler(w http.ResponseWriter, r *http.Request) {
	u, err := url.Parse(r.URL.Query().Get("u"))
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	msg := PersonalURL{URI: u.String(), IP: remoteIP(r)}
	enc, err := encode(msg)
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
	dec, err := decode(vars["enc"])
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	if rip := remoteIP(r); dec.IP != rip {
		log.Print(dec.IP, rip)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, dec.URI, http.StatusFound)
	return
}

func main() {
	router.Handle("/favicon.ico", http.NotFoundHandler())
	router.HandleFunc("/", EncodeHandler).Methods("GET")
	router.HandleFunc("/{enc}", DecodeHandler).Methods("GET").Name("Decode")
	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":7001", nil))
}
