package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gtfierro/savepoint/api"
	bw "gopkg.in/immesys/bw2bind.v5"
	"log"
	"os"
	"strings"
)

func applyArchive(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.String("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	uri := strings.TrimSuffix(c.String("uri"), "/")
	request := api.ArchiveRequest{
		URI:             c.String("uri"),
		PO:              bw.FromDotForm(c.String("po")),
		ValueExpr:       c.String("value"),
		TimeExpr:        c.String("time"),
		UUIDExpr:        c.String("uuid"),
		InheritMetadata: c.BoolT("inheritMetadata"),
	}
	err := API.AttachArchiveRequests(uri, &request)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func removeRequest(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.String("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	if c.String("uri") == "" {
		return fmt.Errorf("Need URI")
	}
	uri := strings.TrimSuffix(c.String("uri"), "/")
	err := API.RemoveArchiveRequests(uri, false)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func scanRequests(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.String("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	if c.String("uri") == "" {
		return fmt.Errorf("Need URI")
	}
	uri := strings.TrimSuffix(c.String("uri"), "/")
	requests, err := API.GetArchiveRequests(uri)
	if err != nil {
		log.Fatal(err)
	}
	for _, req := range requests {
		fmt.Println("---------------")
		req.Dump()
	}
	return nil
}

func addConfig(c *cli.Context) error {
	config, err := api.ReadConfig(c.String("config"))
	if err != nil {
		log.Fatal(err)
	}
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.String("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	for _, req := range config.DummyArchiveRequests {
		archiveRequest := req.ToArchiveRequest()
		err := API.AttachArchiveRequests(req.AttachURI, archiveRequest)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func rmConfig(c *cli.Context) error {
	config, err := api.ReadConfig(c.String("config"))
	if err != nil {
		log.Fatal(err)
	}
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.String("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	for _, req := range config.DummyArchiveRequests {
		err := API.RemoveArchiveRequests(req.AttachURI, false)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "savepoint"
	app.Usage = "Utility for managing Archive Requests"
	app.Version = "0.0.6"

	app.Commands = []cli.Command{
		{
			Name:   "archive",
			Usage:  "Request that a given URI be archived",
			Action: applyArchive,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uri,u",
					Usage: "REQUIRED. The URI you want to archive",
				},
				cli.StringFlag{
					Name:  "value,v",
					Value: "key1",
					Usage: "REQUIRED. The objectbuilder expression for where the value is in messages published on the URI (see https://github.com/gtfierro/ob)",
				},
				cli.StringFlag{
					Name:  "time,t",
					Value: "",
					Usage: "OPTIONAL. Objectbuilder expression for where the timestamp is",
				},
				cli.StringFlag{
					Name:  "uuid",
					Value: "",
					Usage: "OPTIONAL. Objectbuilder expression for where the uuid is",
				},
				cli.StringFlag{
					Name:  "po",
					Value: "2.0.0.0",
					Usage: "OPTIONAL (uses default). The kind of PO to parse for applying the expression",
				},
				cli.StringFlag{
					Name:  "parse,p",
					Value: "2006-01-02T15:04:05Z07:00",
					Usage: "OPTIONAL. How to parse the timestamp",
				},
				cli.BoolTFlag{
					Name:  "inheritMetadata,i",
					Usage: "OPTIONAL. Defaults to true. Inherits metadata from all URI prefixes",
				},
				cli.StringFlag{
					Name:   "entity,e",
					EnvVar: "BW2_DEFAULT_ENTITY",
					Usage:  "The entity to use",
				},
			},
		},
		{
			Name:   "addc",
			Usage:  "Load archive requests from config file",
			Action: addConfig,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config,c",
					Value: "archive.yml",
					Usage: "Config file to parse for archive requests",
				},
				cli.StringFlag{
					Name:   "entity,e",
					EnvVar: "BW2_DEFAULT_ENTITY",
					Usage:  "The entity to use",
				},
			},
		},
		{
			Name:   "rmc",
			Usage:  "Remove archive requests identified by config file",
			Action: rmConfig,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "config,c",
					Value: "archive.yml",
					Usage: "Config file to parse for archive requests",
				},
				cli.StringFlag{
					Name:   "entity,e",
					EnvVar: "BW2_DEFAULT_ENTITY",
					Usage:  "The entity to use",
				},
			},
		},
		{
			Name:   "remove",
			Usage:  "Request that a URI stop being archived (does not delete data)",
			Action: removeRequest,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uri,u",
					Usage: "URI to remove metadata !meta/giles from",
				},
				cli.StringFlag{
					Name:   "entity,e",
					EnvVar: "BW2_DEFAULT_ENTITY",

					Usage: "The entity to use",
				},
			},
		},
		{
			Name:   "scan",
			Usage:  "Scan for archive requests",
			Action: scanRequests,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "uri,u",
					Usage: "Base URI to scan for metadata matching !meta/giles",
				},
				cli.StringFlag{
					Name:   "entity,e",
					EnvVar: "BW2_DEFAULT_ENTITY",
					Usage:  "The entity to use",
				},
			},
		},
	}
	app.Run(os.Args)
}
