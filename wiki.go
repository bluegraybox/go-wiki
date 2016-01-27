/*
Simple file-backed wiki with markdown support
*/

package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/shurcooL/github_flavored_markdown"
	ht "html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	tt "text/template"
	// "github.com/microcosm-cc/bluemonday"
	// "github.com/russross/blackfriday"
)

/******************************************************************************/
/*                          Page struct and methods                           */
/******************************************************************************/

type pageIO struct {
	PagesDir string
}

type Page struct {
	Title string
	Body  []byte
}

func (pi *pageIO) save(p *Page) error {
	filename := pi.pageFile(p.Title)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func (p *pageIO) loadPage(title string) (*Page, error) {
	body, err := ioutil.ReadFile(p.pageFile(title))
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func (p *pageIO) initPagesDir() {
	err := os.MkdirAll(p.PagesDir, 0700)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (p *pageIO) pageFile(title string) string {
	return p.PagesDir + "/" + title + ".txt"
}

/******************************************************************************/
/*                          HTTP server and handlers                          */
/******************************************************************************/

const viewPrefix = len("/view/")
const editPrefix = len("/edit/")
const savePrefix = len("/save/")
const renamePrefix = len("/rename/")

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
	return &config{os.Getenv("username"), os.Getenv("password")}
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(p pageIO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[viewPrefix:] // trim leading '/view/'
		page, err := p.loadPage(title)
		if err != nil {
			http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		} else {
			renderMarkdown(page, w, "view.html")
		}
	}
}

func editHandler(p pageIO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[editPrefix:] // trim leading '/edit/'
		page, err := p.loadPage(title)
		if err != nil {
			page = &Page{Title: title}
		}
		renderHtml(page, w, "edit.html")
	}
}

func saveHandler(p pageIO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[savePrefix:] // trim leading '/save/'
		body := r.FormValue("body")
		// fmt.Printf("saveHandler body: %q\n", body)
		page := &Page{Title: title, Body: []byte(body)}
		err := p.save(page)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		} else {
			http.Redirect(w, r, "/view/"+title, http.StatusFound)
		}
	}
}

func renameHandler(p pageIO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		title := r.URL.Path[renamePrefix:] // trim leading '/rename/'
		if r.Method == "GET" {
			page := &Page{Title: title}
			renderHtml(page, w, "rename.html")
		} else if r.Method == "POST" {
			newName := r.FormValue("newName")
			oldFilename := p.pageFile(title)
			newFilename := p.pageFile(newName)
			if _, err := os.Stat(oldFilename); os.IsNotExist(err) {
				// If the old file doesn't exist, we still want to rewrite the links,
				// and redirect to the edit page for the new name.
				err = p.rewriteLinks(title, newName)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					http.Redirect(w, r, "/edit/"+newName, http.StatusFound)
				}
			} else {
				err := os.Rename(oldFilename, newFilename)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				} else {
					err = p.rewriteLinks(title, newName)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					} else {
						http.Redirect(w, r, "/view/"+newName, http.StatusFound)
					}
				}
			}
		} else {
			http.Error(w, fmt.Sprintf("Invalid Method '%s'", r.Method), http.StatusMethodNotAllowed)
		}
	}
}

func (p pageIO) rewriteLinks(title, newName string) error {
	newFilename := p.pageFile(newName)
	files, err := ioutil.ReadDir(p.PagesDir)
	if err != nil {
		return err
	} else {
		renameRegex := regexp.MustCompile(`(\[[^\]]*\])\(` + title + `\)`)
		for _, f := range files {
			filename := p.PagesDir + "/" + f.Name()
			if filename != newFilename {
				content, err := ioutil.ReadFile(filename)
				if err != nil {
					return err
				} else {
					newContent := renameRegex.ReplaceAll(content, []byte("$1("+newName+")"))
					err = ioutil.WriteFile(filename, newContent, 0644)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func allHandler(p pageIO) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		files, err := ioutil.ReadDir(p.PagesDir)
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
}

func baseName(info os.FileInfo) string {
	l := len(info.Name())
	if info.Name()[l-4:] == ".txt" {
		return info.Name()[0 : l-4]
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
	portPtr := flag.Int("port", 80, "HTTP server port")
	pagesDirPtr := flag.String("pages", "/var/local/wiki", "Directory to store wiki pages in")
	flag.Parse()
	port := fmt.Sprintf(":%d", *portPtr)

	pIO := pageIO{*pagesDirPtr}

	pIO.initPagesDir()
	c := loadConfig()
	secWrap := makeSecWrap(c)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/view/", secWrap(viewHandler(pIO)))
	http.HandleFunc("/edit/", secWrap(editHandler(pIO)))
	http.HandleFunc("/save/", secWrap(saveHandler(pIO)))
	http.HandleFunc("/rename/", secWrap(renameHandler(pIO)))
	http.HandleFunc("/all/", secWrap(allHandler(pIO)))
	http.HandleFunc("/", secWrap(defaultHandler))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
