package main

import (
	"code.google.com/p/go-sqlite/go1/sqlite3"
	"fmt"
	"hash/fnv"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
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
	if err != nil {
		renderError(w, r, "Database error!", err)
		return
	}
	defer connection.Close()

	url, err := url.Parse(link)
	if err != nil {
		renderError(w, r, "Invalid URL", err)
		return
	}

	//if the website uses http(s), append http
	if len(url.Scheme) == 0 {
		link = "http://" + link
	}

	//check if this link is already in the database
	sql := fmt.Sprintf("SELECT hash FROM url_mapping WHERE url = \"%s\"", link)
	hash_query, err := connection.Query(sql)

	//if not found...
	if err == io.EOF && hash_query == nil {
		h := hash(link)
		sql := fmt.Sprintf("INSERT INTO url_mapping VALUES(NULL, \"%s\", \"%s\");", h, link)
		err = connection.Exec(sql)
		if err != nil {
			renderError(w, r, "Failed to update database!", err)
			return
		}
		txt := fmt.Sprintf("%s mapped to %s", link, h)
		msg := &Message{Text: txt, Type: "success"}
		renderTemplate(w, r, msg)
		return
	}
	defer hash_query.Close()

	// already in DB, return hash that exists
	var return_hash string
	//wtf???
	hash_query.Scan(&return_hash)
	txt := fmt.Sprintf("%s mapped to %s", link, return_hash)
	msg := &Message{Text: txt, Type: "success"}
	renderTemplate(w, r, msg)
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	link, err := getLink(w, r)
	if err != nil {
		renderError(w, r, "URL parse error!", err)
		return
	}

	if len(link) == 0 {
		renderTemplate(w, r, nil)
		return
	}

	connection, err := sqlite3.Open("src/tinyurl/url.db")
	if err != nil {
		renderError(w, r, "Database error!", err)
		return
	}
	defer connection.Close()

	sql := fmt.Sprintf("SELECT url FROM url_mapping WHERE hash = \"%s\"", link)
	url_query, err := connection.Query(sql)
	if err != nil {
		renderError(w, r, "That tinyurl link does not exist!", err)
		return
	}
	defer url_query.Close()

	var url string
	url_query.Scan(&url)
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func renderError(w http.ResponseWriter, r *http.Request, text string, err error) {
	fmt.Printf(err.Error())
	msg := &Message{Text: text, Type: "error"}
	renderTemplate(w, r, msg)
	return
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
