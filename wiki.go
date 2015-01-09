package main

import (
    "fmt"
    "net/http"
    "io/ioutil"
)


/******************************************************************************/
/*                          Page struct and methods                           */
/******************************************************************************/

type Page struct {
    Title string
    Body []byte
}

func (p *Page) save() error {
    filename := pageFile(p.Title)
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
    body, err := ioutil.ReadFile(pageFile(title))
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func pageFile(title string) string {
    return title + ".txt"
}

/******************************************************************************/
/*                          HTTP server and handlers                          */
/******************************************************************************/

func handler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "message: %s", r.URL.Path[1:])  // trim leading '/'
}

func wikiHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[6:]  // trim leading '/wiki/'
    page, err := loadPage(title)
    if err != nil {
        w.WriteHeader(404)
        fmt.Fprintf(w, "Error: %s", err)
    } else {
        fmt.Fprintf(w, "<h1>%s</h1><p>%s</p>", page.Title, page.Body)
    }
}

func main() {
    http.HandleFunc("/wiki/", wikiHandler)
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8888", nil)
}
