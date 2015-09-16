package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"hash/fnv"
	"strconv"
	"code.google.com/p/go-sqlite/go1/sqlite3"
)

type Message struct {
	Text string
	Type string
}

var validPath = regexp.MustCompile("^/(save|([a-zA-Z0-9]*))$")

func saveUrlHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	link := r.Form.Get("url-input")

	connection, err := sqlite3.Open("src/tinyurl/url.db")
	defer connection.Close()
	if err != nil {
		msg := &Message{Text: "Database error!", Type: "error"}
		renderTemplate(w, r, msg)
		return
	}

	sql := fmt.Sprintf("SELECT url FROM url_mapping WHERE url = \"%s\"", link)
	_, err = connection.Query(sql)
	if err != nil {
		url, err := url.Parse(link)
		if err != nil {
			msg := &Message{Text: "Invalid URL", Type: "Error"}
			renderTemplate(w, r, msg)
		}
		if len(url.Scheme) == 0 {
			link = "http://" + link
		}
		h := hash(link)
		sql := fmt.Sprintf("INSERT INTO url_mapping VALUES(%s, %s);", h, link)
		_, err = connection.Query(sql)
		if err != nil {
			msg := &Message{Text: "Failed to update database!", Type: "error"}
			renderTemplate(w, r, msg)
			return
		}
		txt := fmt.Sprintf("%s mapped to %s!", link, h)
		msg := &Message{Text: txt, Type: "success"}
		renderTemplate(w, r, msg)
		return
	}
	// already in DB, return hash that exists
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	link, err := getLink(w, r)
	if err != nil {
		return
	}
	if len(link) == 0 {
		renderTemplate(w, r, nil)
		return
	}

	connection, err := sqlite3.Open("src/tinyurl/url.db")
	defer connection.Close()
	if err != nil {
		msg := &Message{Text: "Database error!", Type: "error"}
		renderTemplate(w, r, msg)
		return
	}

	sql := fmt.Sprintf("SELECT url FROM url_mapping WHERE hash = \"%s\"", link)
	url_query, err := connection.Query(sql)
	if err != nil {
		msg := &Message{Text: "That tinyurl link does not exist!", Type: "error"}
		renderTemplate(w, r, msg)
		return
	}

	var url string
	url_query.Scan(&url)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func renderTemplate(w http.ResponseWriter, r *http.Request, msg *Message) {
	tmpl, err := template.ParseFiles("src/tinyurl/home.html")
	if err != nil {
		http.NotFound(w, r)
	}
	tmpl.Execute(w, msg)
	return
}

func getLink(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", nil
	}
	return m[1], nil
}

func hash(link string) string {
	return strconv.FormatUint(uint64(createHash(link)), 16)
}

func createHash(link string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(link))
	return h.Sum32()
}

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/save", saveUrlHandler)
	http.ListenAndServe(":8998", nil)
}
