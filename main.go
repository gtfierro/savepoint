package main

import (
	"./api"
	"github.com/codegangsta/cli"
	bw "gopkg.in/immesys/bw2bind.v5"
	"log"
	"os"
	"strings"
)

func applyArchive(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.GlobalString("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	uri := strings.TrimSuffix(c.String("uri"), "/")
	request := api.ArchiveRequest{
		URI:         c.String("uri"),
		PO:          bw.FromDotForm(c.String("po")),
		Value:       c.String("value"),
		Time:        c.String("time"),
		UUID:        c.String("uuid"),
		MetadataURI: c.String("metadataURI"),
	}
	err := API.AttachArchiveRequests(uri, &request)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func removeRequest(c *cli.Context) error {
	client := bw.ConnectOrExit("")
	vk := client.SetEntityFileOrExit(c.GlobalString("entity"))
	client.OverrideAutoChainTo(true)
	API := api.NewAPI(client, vk)
	uri := strings.TrimSuffix(c.String("uri"), "/")
	err := API.RemoveArchiveRequests(uri, false)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "savepoint"
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "entity,e",
			Value:  "",
			Usage:  "The entity to use",
			EnvVar: "BW2_DEFAULT_ENTITY",
		},
	}

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
				cli.StringFlag{
					Name:  "metadataURI,mu",
					Value: "",
					Usage: "OPTIONAL. Specifies base uri <uri>/!meta/+ for metadata keys",
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
			},
		},
	}
	app.Run(os.Args)
}
