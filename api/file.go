package api

import (
	"fmt"
	"github.com/pkg/errors"
	bw "gopkg.in/immesys/bw2bind.v5"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
)

// Struct representation of a configuration file for attaching archive
// metadata. Follows the basic structure:
//
//    Prefix: gabe.pantry/services
//    Archive:
//      - URI: s.TED/MTU1/i.meter/signal/Voltage
//        Value: Value
//      - URI: s.TED/MTU1/i.meter/signal/PowerNow
//        Value: Value
//      - URI: s.TED/MTU1/i.meter/signal/KVA
//        Value: Value
type Config struct {
	Prefix               string                `yaml:"Prefix"`
	DummyArchiveRequests []DummyArchiveRequest `yaml:"Archive"`
	//ArchiveRequests      []*ArchiveRequest
}

type DummyArchiveRequest struct {
	AttachURI       string   `yaml:"AttachURI"`
	ArchiveURI      string   `yaml:"ArchiveURI"`
	PO              string   `yaml:"PO"`
	UUID            string   `yaml:"UUID"`
	Value           string   `yaml:"Value"`
	Time            string   `yaml:"Time"`
	TimeParse       string   `yaml:"TimeParse"`
	InheritMetadata string   `yaml:"InheritMetadata",omitempty`
	MetadataURIs    []string `yaml:"MetadataURIs"`
	MetadataBlock   string   `yaml:"MetadataBlock"`
	MetadataExpr    string   `yaml:"MetadataExpr"`
}

func (d DummyArchiveRequest) ToArchiveRequest() *ArchiveRequest {
	var doinherit = true
	if d.InheritMetadata == "false" {
		doinherit = false
	}
	req := &ArchiveRequest{
		URI:             d.ArchiveURI,
		PO:              bw.FromDotForm(d.PO),
		UUID:            d.UUID,
		Value:           d.Value,
		Time:            d.Time,
		TimeParse:       d.TimeParse,
		InheritMetadata: doinherit,
		MetadataURIs:    d.MetadataURIs,
		MetadataBlock:   d.MetadataBlock,
		MetadataExpr:    d.MetadataExpr,
	}

	if d.AttachURI == "" {
		d.AttachURI = d.ArchiveURI
	}

	return req
}

func ReadConfig(filename string) (*Config, error) {
	fmt.Printf("Reading from config file %s\n", filename)
	var config = new(Config)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, errors.Wrap(err, fmt.Sprintf("Could not read config file %v", filename))
	}
	if err := yaml.Unmarshal(bytes, config); err != nil {
		return config, errors.Wrap(err, "Could not unmarshal config file")
	}
	config.Prefix = strings.TrimSuffix(config.Prefix, "/")

	if len(config.DummyArchiveRequests) == 0 {
		return config, errors.New("Need to provide archive requests")
	}
	for i, req := range config.DummyArchiveRequests {
		req.ArchiveURI = config.Prefix + "/" + strings.TrimPrefix(req.ArchiveURI, "/")
		req.AttachURI = config.Prefix + "/" + strings.TrimPrefix(req.AttachURI, "/")
		if req.PO == "" {
			req.PO = "2.0.0.0"
		}
		for idx, uri := range req.MetadataURIs {
			req.MetadataURIs[idx] = config.Prefix + "/" + strings.TrimPrefix(uri, "/")
		}
		config.DummyArchiveRequests[i] = req
	}

	return config, nil
}
