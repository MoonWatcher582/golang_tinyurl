package main

import (
	"fmt"
	"html/template"
	//"io/ioutil"
	"net/http"
	"regexp"
	"code.google.com/p/go-sqlite/go1/sqlite3"
)

var validPath = regexp.MustCompile("^/(save|([a-zA-Z0-9]*))$")

func saveUrlHandler(w http.ResponseWriter, r *http.Request) {
	//hash url
	//store hash and url
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	var url string;
	link, err := getLink(w, r)
	if err != nil {
		return
	}
	if len(link) == 0 {
		tmpl, err := template.ParseFiles("home.html")
		if err != nil {
			http.NotFound(w, r)
		}
		tmpl.Execute(w, nil)
		return
	}
	connection, err := sqlite3.Open("url.db")
	defer connection.Close()
	if err != nil {
		fmt.Fprintf(w, "DB Error")
	}

	sql := fmt.Sprintf("SELECT url FROM url_mapping WHERE hash = \"%s\"", link)
	url_query, err := connection.Query(sql)
	if err != nil {
		fmt.Printf("Could not find this hash")
		http.Redirect(w, r, "/", http.StatusOK)
	}
	url_query.Scan(url)
	fmt.Printf("URL: %s", url)
	http.Redirect(w, r, url, http.StatusOK)
}

func getLink(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", nil
	}
	return m[1], nil
}

func main() {
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/save", saveUrlHandler)
	http.ListenAndServe(":8998", nil)
}
