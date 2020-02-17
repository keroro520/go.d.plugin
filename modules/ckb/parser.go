package ckb

import (
	"os"
	"encoding/json"
	"bufio"
	"strings"
	"errors"
)

type Metric struct {
	Measurement string `json:"measurement"`
	Tags map[string]string `json:"tags"`
	Fields map[string]uint64 `json:"fields"`
}

type Parser struct {
	file *os.File
	reader *bufio.Reader
}

func NewParser(file *os.File) *Parser {
	reader := bufio.NewReader(file)
	parser := Parser{
		file: file,
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
		return nil, errors.New("not a metric line")
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
		p.reader = nil
		p.file = nil
	}
}
