/*
Demonstrate template error conditions.
*/
package main

import (
    "bytes"
    "html/template"
    "testing"
)


type Thing struct {
    Name string
}

func TestTemplate(t *testing.T) {
    thing := &Thing{Name: "one"}
    var out bytes.Buffer
    tmpl, err := template.New("broken").Parse(`<p>{{printf "%s" .UnknownAttribute}}</p>`)
    // should not get an error parsing the template...
    if err != nil {
        t.Errorf("err: %v", err)
    } else {
        // but should get an error executing it
        err = tmpl.Execute(&out, thing)
        if err == nil {
            t.Error("expected error executing bad template")
        }
    }
}
