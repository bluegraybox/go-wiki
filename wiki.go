/*
Simple file-backed wiki with markdown support
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/shurcooL/go/github_flavored_markdown"
	ht "html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	tt "text/template"
	// "github.com/microcosm-cc/bluemonday"
	// "github.com/russross/blackfriday"
)

/******************************************************************************/
/*                          Page struct and methods                           */
/******************************************************************************/

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	initPagesDir()
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

const pagesDir = "/var/local/wiki"

func initPagesDir() {
	err := os.MkdirAll(pagesDir, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func pageFile(title string) string {
	return pagesDir + "/" + title + ".txt"
}

/******************************************************************************/
/*                          HTTP server and handlers                          */
/******************************************************************************/

const viewPrefix = len("/view/")
const editPrefix = len("/edit/")
const savePrefix = len("/save/")

type handler func(http.ResponseWriter, *http.Request)

func setAuthRespHeader(w http.ResponseWriter) {
	w.Header().Add("WWW-Authenticate", `Basic realm="wiki"`)
}

func setBadHeaderResponse(w http.ResponseWriter) {
	setAuthRespHeader(w)
	http.Error(w, "Authorization header should be \"Basic username:password\"", http.StatusUnauthorized)
}

func makeSecWrap(c *config) func(handler) handler {
	validate := func(username, password string) bool {
		if username == c.Username && password == c.Password {
			return true
		}
		return false
	}
	return func(f handler) handler {
		return func(w http.ResponseWriter, r *http.Request) {
			// Core logic from http://bl.ocks.org/tristanwietsma/8444cf3cb5a1ac496203
			authHdr := r.Header.Get("Authorization")
			auth := strings.SplitN(authHdr, " ", 2)
			if len(auth) != 2 {
				setBadHeaderResponse(w)
				return
			}
			if auth[0] != "Basic" {
				setBadHeaderResponse(w)
				return
			}

			userPwd, err := base64.StdEncoding.DecodeString(auth[1])
			if err != nil {
				setAuthRespHeader(w)
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			pair := strings.SplitN(string(userPwd), ":", 2)
			if len(pair) != 2 {
				setBadHeaderResponse(w)
				return
			}

			if !validate(pair[0], pair[1]) {
				setAuthRespHeader(w)
				http.Error(w, "Username/password validation failed", http.StatusUnauthorized)
				return
			}

			f(w, r)
		}
	}
}

type config struct {
	Username string
	Password string
}

func loadConfig() *config {
	// We want this to be fragile
	body, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Fatal(err)
	}
	c := &config{}
	json.Unmarshal(body, c)
	return c
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[viewPrefix:] // trim leading '/view/'
	page, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
	} else {
		renderMarkdown(page, w, "view.html")
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[editPrefix:] // trim leading '/edit/'
	page, err := loadPage(title)
	if err != nil {
		page = &Page{Title: title}
	}
	renderHtml(page, w, "edit.html")
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[savePrefix:] // trim leading '/save/'
	body := r.FormValue("body")
	// fmt.Printf("saveHandler body: %q\n", body)
	page := &Page{Title: title, Body: []byte(body)}
	err := page.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
	}
}

func allHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(pagesDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	t, err := ht.ParseFiles("all.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		var out bytes.Buffer
		baseNames := make([]string, len(files))
		for i, f := range files {
			baseNames[i] = baseName(f)
		}
		err := t.Execute(&out, baseNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write(out.Bytes())
		}
	}
}

func baseName(info os.FileInfo) string {
	l := len(info.Name())
	if info.Name()[l-4:] == ".txt" {
		return info.Name()[0:l-4]
	}
	return info.Name()
}

type tmpl interface {
	Execute(out io.Writer, data interface{}) error
}

type renderer func([]byte) []byte
type templateParser func(string) (tmpl, error)

func renderHtml(p *Page, w http.ResponseWriter, tFile string) {
	// htmlParser := ht.ParseFiles  // this doesn't work
	htmlParser := func(f string) (t tmpl, e error) { return ht.ParseFiles(f) }
	noopRenderer := func(body []byte) []byte { return body }
	renderPage(p, w, tFile, htmlParser, noopRenderer)
}

func renderMarkdown(p *Page, w http.ResponseWriter, tFile string) {
	textParser := func(f string) (t tmpl, e error) { return tt.ParseFiles(f) }
	markdownRenderer := func(body []byte) []byte {
		var rendered bytes.Buffer
		rendered.Write(github_flavored_markdown.Markdown(body))
		// html := bluemonday.UGCPolicy().SanitizeBytes(p.Body)
		// rendered.Write(blackfriday.MarkdownCommon(html))
		return rendered.Bytes()
	}
	renderPage(p, w, tFile, textParser, markdownRenderer)
}

func renderPage(p *Page, w http.ResponseWriter, tFile string, parser templateParser, renderFunc renderer) {
	t, err := parser(tFile)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		p.Body = renderFunc(p.Body)
		// if we t.Execute directly to w, then http.Error doesn't set the status code,
		// and the template up to the point of failure will still be in the output body
		var out bytes.Buffer
		err := t.Execute(&out, p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			w.Write(out.Bytes())
		}
	}
}

func main() {
	initPagesDir()
	c := loadConfig()
	secWrap := makeSecWrap(c)
	http.HandleFunc("/view/", secWrap(viewHandler))
	http.HandleFunc("/edit/", secWrap(editHandler))
	http.HandleFunc("/save/", secWrap(saveHandler))
	http.HandleFunc("/all/", secWrap(allHandler))
	http.HandleFunc("/", secWrap(defaultHandler))
	http.ListenAndServe(":80", nil)
}
