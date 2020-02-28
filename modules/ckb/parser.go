package ckb

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"github.com/coreos/go-systemd/sdjournal"
)

type Metric struct {
	Topic  string            `json:"topic"`
	Tags   map[string]string `json:"tags"`
	Fields map[string]uint64 `json:"fields"`
}

type Parser struct {
	file   *os.File
	reader *bufio.Reader
	journal *sdjournal.JournalReader
}

func NewFileParser(file *os.File) *Parser {
	reader := bufio.NewReader(file)
	parser := Parser{
		file:   file,
		reader: reader,
	}
	return &parser
}

func NewJournalParser(journal *sdjournal.JournalReader) *Parser {
	reader := bufio.NewReader(journal)
	parser := Parser{
		journal: journal,
		reader: reader,
	}
	return &parser
}

func (p *Parser) ReadLine() (*Metric, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	index := strings.Index(line, "ckb-metrics ")
	if index == -1 {
		return nil, nil
	}

	var metric Metric
	err = json.Unmarshal([]byte(line[index+12:]), &metric)
	if err != nil {
		return nil, err
	}

	return &metric, nil
}

func (p *Parser) Close() {
	if p.file != nil {
		p.file.Close()
		p.file = nil
		p.reader = nil
	}
	if p.journal != nil {
		p.journal.Close()
		p.journal = nil
		p.reader = nil
	}
}
