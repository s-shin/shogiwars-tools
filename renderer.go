package main

import (
	"io"
	"strings"
)

type ASCIIRenderer interface {
	SetHeader(cols []string)
	Append(cols []string)
	Render()
}

type TsvRenderer struct {
	w      io.Writer
	header []string
	rows   [][]string
}

func NewTsvRenderer(w io.Writer) *TsvRenderer {
	return &TsvRenderer{
		w:      w,
		header: nil,
		rows:   make([][]string, 0),
	}
}

func (r *TsvRenderer) SetHeader(cols []string) {
	r.header = cols
}

func (r *TsvRenderer) Append(cols []string) {
	r.rows = append(r.rows, cols)
}

func (r *TsvRenderer) Render() {
	if r.header != nil {
		r.w.Write([]byte(strings.Join(r.header, "\t")))
		r.w.Write([]byte("\n"))
	}
	for _, cols := range r.rows {
		r.w.Write([]byte(strings.Join(cols, "\t")))
		r.w.Write([]byte("\n"))
	}
}
