package main

import (
    "bytes"
    "net/http"
    "io/ioutil"
    "html/template"
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

const viewPrefix = len("/view/")
const editPrefix = len("/edit/")
const savePrefix = len("/save/")

func handler(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, "/view/FrontPage", http.StatusFound)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[viewPrefix:]  // trim leading '/view/'
    page, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
    } else {
        renderPage(page, w, "view.html")
    }
}

func editHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[editPrefix:]  // trim leading '/edit/'
    page, err := loadPage(title)
    if err != nil {
        page = &Page{Title: title}
    }
    renderPage(page, w, "edit.html")
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
    title := r.URL.Path[savePrefix:]  // trim leading '/save/'
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

func renderPage(p *Page, w http.ResponseWriter, tFile string) {
    t, err := template.ParseFiles(tFile)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    } else {
        var out bytes.Buffer
        err = t.Execute(&out, p)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
        } else {
            w.Write(out.Bytes())
        }
    }
}

func main() {
    http.HandleFunc("/view/", viewHandler)
    http.HandleFunc("/edit/", editHandler)
    http.HandleFunc("/save/", saveHandler)
    http.HandleFunc("/", handler)
    http.ListenAndServe(":8888", nil)
}
