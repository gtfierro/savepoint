package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gtfierro/savepoint/api"
	bw "gopkg.in/immesys/bw2bind.v5"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func getEntityFile(filename string) string {
	if filename == "~/.savepoint.ent" {
		user, err := user.Current()
		if err != nil {
			log.Fatal(err)
		}
		return filepath.Join(user.HomeDir, ".savepoint.ent")
	}
	return filename
}

func applyArchive(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(getEntityFile(c.GlobalString("entity")))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	uri := strings.TrimSuffix(c.String("uri"), "/")
	request := api.ArchiveRequest{
		URI:          c.String("uri"),
		PO:           bw.FromDotForm(c.String("po")),
		Value:        c.String("value"),
		Time:         c.String("time"),
		UUID:         c.String("uuid"),
		MetadataURIs: c.StringSlice("metadataURI"),
	}
	err := API.AttachArchiveRequests(uri, &request)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func removeRequest(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(getEntityFile(c.GlobalString("entity")))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	uri := strings.TrimSuffix(c.String("uri"), "/")
	err := API.RemoveArchiveRequests(uri, false)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func scanRequests(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(getEntityFile(c.GlobalString("entity")))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
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
	vk := client.SetEntityFileOrExit(getEntityFile(c.GlobalString("entity")))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	for _, req := range config.ArchiveRequests {
		err := API.AttachArchiveRequests(req.URI, req)
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
	vk := client.SetEntityFileOrExit(getEntityFile(c.GlobalString("entity")))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	for _, req := range config.ArchiveRequests {
		err := API.RemoveArchiveRequests(req.URI, false)
		if err != nil {
			log.Fatal(err)
		}
	}
	return nil
}

func setDefaultEntity(c *cli.Context) error {
	user, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}
	entityFile := c.String("entity")
	file, err := os.Create(filepath.Join(user.HomeDir, ".savepoint.ent"))
	if err != nil {
		log.Fatal(err)
	}
	entity, err := os.Open(entityFile)
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(file, entity)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "savepoint"
	app.Version = "0.0.3"

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
					Usage: "REQUIRED. The objectbuilder expression for where the value is in messages published on the URI (see https://github.com/gtfierro/giles2/tree/master/objectbuilder)",
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
				cli.StringSliceFlag{
					Name:  "metadataURI,mu",
					Usage: "OPTIONAL. Specifies base uri <uri>/!meta/+ for metadata keys",
				},
				cli.StringFlag{
					Name:  "entity,e",
					Value: "~/.savepoint.ent",
					Usage: "The entity to use",
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
					Name:  "entity,e",
					Value: "~/.savepoint.ent",
					Usage: "The entity to use",
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
					Name:  "entity,e",
					Value: "~/.savepoint.ent",
					Usage: "The entity to use",
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
					Value: "",
					Usage: "URI to remove metadata !meta/giles from",
				},
				cli.StringFlag{
					Name:  "entity,e",
					Value: "~/.savepoint.ent",
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
					Value: "scratch.ns/*",
					Usage: "Base URI to scan for metadata matching !meta/giles",
				},
				cli.StringFlag{
					Name:  "entity,e",
					Value: "~/.savepoint.ent",
					Usage: "The entity to use",
				},
			},
		},
		{
			Name:   "setEntity",
			Usage:  "Set default entity for this utility to use",
			Action: setDefaultEntity,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:   "entity,e",
					Value:  "",
					Usage:  "The entity to use",
					EnvVar: "BW2_DEFAULT_ENTITY",
				},
				cli.StringFlag{
					Name:  "entity,e",
					Value: "~/.savepoint.ent",
					Usage: "The entity to use",
				},
			},
		},
	}
	app.Run(os.Args)
}
