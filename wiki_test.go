package main

import (
    "strings"
    "testing"
    "net/http"
    "net/http/httptest"
)


func TestSaveLoad(t *testing.T) {
    p1 := &Page{Title: "TestPage", Body: []byte("This is a sample page")}
    p1.save()
    p2, err := loadPage("TestPage")
    if err != nil {
        t.Errorf("Error loading page: %v", err)
    }
    t.Logf("Page body: %v", string(p2.Body))
}

func TestHandler(t *testing.T) {
    response := httptest.NewRecorder()
    request, _ := http.NewRequest("GET", "http://domain.com/just/a/test", nil)
    handler(response, request)
    msg := response.Body.String()
    if msg != "message: just/a/test" {
        t.Errorf("Wrong response: %s", msg)
    }
}

func TestWikiHandler(t *testing.T) {
    p1 := &Page{Title: "TestWikiPage", Body: []byte("This is a sample wiki page")}
    p1.save()

    response := httptest.NewRecorder()
    request, _ := http.NewRequest("GET", "http://domain.com/wiki/TestWikiPage", nil)
    wikiHandler(response, request)
    body := response.Body.String()
    if !(strings.Contains(body, "<h1>TestWikiPage</h1>") &&
            strings.Contains(body, "<p>This is a sample wiki page</p>")) {
        t.Errorf("Wrong response body: %s", body)
    }
}

func TestMissingWikiHandler(t *testing.T) {
    response := httptest.NewRecorder()
    request, _ := http.NewRequest("GET", "http://domain.com/wiki/TestMissingWikiPage", nil)
    wikiHandler(response, request)
    body := response.Body.String()
    if !strings.Contains(body, "Error: open TestMissingWikiPage.txt: no such file or directory") {
        t.Errorf("Wrong response body: %s", body)
    }
}

func TestBadWikiTemplate(t *testing.T) {
    response := httptest.NewRecorder()
    page := &Page{Title: "Bogus", Body: []byte("meaningless")}
    renderPage(page, response, "bogus.html")
    body := response.Body.String()
    if !strings.Contains(body, "Error: open bogus.html: no such file or directory") {
        t.Errorf("Wrong response body: %s", body)
    }
}
