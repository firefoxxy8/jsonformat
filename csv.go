package main

import (
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"
	"text/template"
)

const (
	PATH = iota + 1
	NAME
	FUNC
	PARAM
)

var fieldregexp = regexp.MustCompile(`^\s*([[:alnum:]\.]+)=([[:alnum:]\s\.]+)(?:\s*\|\s*([[:alnum:]]+)(?:\(([[:alnum:]\s]+)\))?)?\s*`)

func CSVCompiler(format string) (Formatter, error) {
	fieldformats := strings.Split(format, ",")
	fieldtemplates, fieldnames := []string{}, []string{}
	for _, fieldformatstring := range fieldformats {
		fieldformat := fieldregexp.FindStringSubmatch(fieldformatstring)
		if fieldformat == nil || len(fieldformat) == 0 {
			return nil, fmt.Errorf("Invalid CSV format string")
		}
		fieldtemplate := "{{." + fieldformat[PATH]
		if len(fieldformat[FUNC]) > 0 {
			fieldtemplate += "|" + fieldformat[FUNC]
			if len(fieldformat[PARAM]) > 0 {
				fieldtemplate += " " + fieldformat[PARAM]
			}
		}
		fieldtemplates = append(fieldtemplates, fieldtemplate+"}}")
		fieldnames = append(fieldnames, fieldformat[NAME])
	}
	fieldtemplate := strings.Join(fieldtemplates, ",")
	t, err := template.New("jsonformat").
		Funcs(defaultFuncs).
		Parse(fieldtemplate)
	if err != nil {
		return nil, err
	}
	return &CSVFormatter{
		fieldnames: fieldnames,
		template:   t,
	}, nil
}

type CSVFormatter struct {
	fieldnames []string
	template   *template.Template
	sync.Once
}

func (c *CSVFormatter) Execute(w io.Writer, data interface{}) error {
	c.Once.Do(func() {
		io.WriteString(w, strings.Join(c.fieldnames, ",")+"\n")
	})
	return c.template.Execute(w, data)
}

func init() {
	Formats["csv"] = Format{
		Compiler:    CSVCompiler,
		Description: `Shortcut for CSV-style output`,
	}
}
