package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

const TEST_PAGE_DIR = "test_pages"

func TestSaveLoad(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	p1 := &Page{Title: "TestPage", Body: []byte("This is a sample page")}
	pIO.save(p1)
	defer os.Remove(TEST_PAGE_DIR + "/TestPage.txt")
	p2, err := pIO.loadPage("TestPage")
	if err != nil {
		t.Errorf("Error loading page: %v", err)
	}
	t.Logf("Page body: %v", string(p2.Body))
}

func TestSaveFail(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	p1 := &Page{Title: "Bad/Page/Name", Body: []byte("Subdirs don't exist")}
	err := pIO.save(p1)
	if err == nil {
		t.Errorf("Successfully saved bogus page?!")
	}
	_, err = pIO.loadPage("Bad/Page/Name")
	if err == nil {
		t.Errorf("Successfully loaded bogus page?!")
	}
}

func TestDefaultHandler(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://domain.com/totally/invalid/path", nil)
	defaultHandler(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Wrong status code: %d", response.Code)
	}
}

func TestViewHandler(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	p1 := &Page{Title: "TestWikiPage", Body: []byte("This is a sample wiki page")}
	pIO.save(p1)
	defer os.Remove(TEST_PAGE_DIR + "/TestWikiPage.txt")

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://domain.com/view/TestWikiPage", nil)
	viewHandler(pIO)(response, request)
	body := response.Body.String()
	if !(strings.Contains(body, ">TestWikiPage</h1>") &&
		strings.Contains(body, ">This is a sample wiki page<")) {
		t.Errorf("Wrong response body: %s", body)
	}
}

func TestMissingViewHandler(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://domain.com/view/TestMissingWikiPage", nil)
	viewHandler(pIO)(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Wrong status code: %d", response.Code)
	}
}

func TestMissingWikiTemplate(t *testing.T) {
	response := httptest.NewRecorder()
	page := &Page{Title: "Whatever", Body: []byte("whatever")}
	renderHtml(page, response, "no_such_template.html")
	body := response.Body.String()
	if !strings.Contains(body, "open no_such_template.html: no such file or directory") {
		t.Errorf("Wrong response body: %s", body)
	}
}

func TestBadWikiTemplate(t *testing.T) {
	response := httptest.NewRecorder()
	page := &Page{Title: "Whatever", Body: []byte("whatever")}
	renderHtml(page, response, "bad_template.html")
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Wrong status code: %d, body:\n%v", response.Code, response.Body.String())
	}
}

func TestEditHandler(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	p1 := &Page{Title: "TestEditPage", Body: []byte("This is a sample wiki page to edit")}
	pIO.save(p1)
	defer os.Remove(TEST_PAGE_DIR + "/TestEditPage.txt")

	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://domain.com/edit/TestEditPage", nil)
	editHandler(pIO)(response, request)
	body := response.Body.String()
	if !(strings.Contains(body, ">Editing TestEditPage</h1>") &&
		strings.Contains(body, ">This is a sample wiki page to edit</textarea>")) {
		t.Errorf("Wrong response body: %s", body)
	}
}

func TestNewPageHandler(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	defer os.Remove(TEST_PAGE_DIR + "/TestNewPage.txt")
	response := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "http://domain.com/view/TestNewPage", nil)
	editHandler(pIO)(response, request)
	body := response.Body.String()
	if !(strings.Contains(body, ">Editing TestNewPage</h1>") &&
		strings.Contains(body, "</textarea>")) {
		t.Errorf("Wrong response body: %s", body)
	}
}

func TestSaveHandler(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	defer os.Remove(TEST_PAGE_DIR + "/TestNewPage.txt")
	response := httptest.NewRecorder()
	content := "New page content"
	request := newPostRequest("http://domain.com/save/TestNewPage", content)
	saveHandler(pIO)(response, request)
	if response.Code != http.StatusFound {
		t.Errorf("Wrong status code: %d", response.Code)
	}
	page, err := pIO.loadPage("TestNewPage")
	if err != nil {
		t.Errorf("Error loading page: %s", err)
	}
	if string(page.Body) != content {
		t.Errorf("Wrong response body: %s", page.Body)
	}
}

func TestSaveHandlerBadTitle(t *testing.T) {
	pIO := pageIO{TEST_PAGE_DIR}
	pIO.initPagesDir()

	response := httptest.NewRecorder()
	request := newPostRequest("http://domain.com/save/Bad/Page/Name", "filler")
	saveHandler(pIO)(response, request)
	if response.Code != http.StatusInternalServerError {
		t.Errorf("Wrong status code: %d", response.Code)
	}
}

func newPostRequest(reqUrl string, content string) *http.Request {
	form := url.Values{}
	form.Set("body", content)
	request, _ := http.NewRequest("POST", reqUrl, strings.NewReader(form.Encode()))
	// Can't parse the form data if these two aren't set
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Add("Content-Length", fmt.Sprintf("%d", len(form.Encode())))
	return request
}
