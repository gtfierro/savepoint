# SavePoint

API and command line tool for archiving data produced by BOSSWAVE drivers and services.

```bash
NAME:
   savepoint - A new cli application

USAGE:
   savepoint [global options] command [command options] [arguments...]

VERSION:
   0.0.6

COMMANDS:
     archive  Request that a given URI be archived
     addc     Load archive requests from config file
     rmc      Remove archive requests identified by config file
     remove   Request that a URI stop being archived (does not delete data)
     scan     Scan for archive requests
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

You will likely use `scan`,`addc` and `rmc` most often

Archiving is done by attaching _archive requests_ in the form of persisted metadata messages on URIs producing data you want to archive.
Currently only timeseries data is supported.


An archive request has the following structure:

```go
type ArchiveRequest struct {
		// AUTOPOPULATED. The entity that requested the URI to be archived.
		FromVK string
     	// OPTIONAL. the URI to subscribe to. Requires building a chain on the URI
     	// from the .FromVK field. If not provided, uses the base URI of where this
     	// ArchiveRequest was stored. For example, if this request was published
     	// on <uri>/!meta/giles, then if the URI field was elided it would default
     	// to <uri>.
     	URI string

     	// Extracts objects of the given Payload Object type from all messages
     	// published on the URI. If elided, operates on all PO types.
     	PO int

     	// OPTIONAL. If provided, this is used as the stream UUID.  If not
     	// provided, then a UUIDv3 with NAMESPACE_UUID and the URI, PO type and
     	// Value are used.
     	UUID string

     	// expression determining how to extract the value from the received
     	// message
     	Value string

     	// OPTIONAL. Expression determining how to extract the value from the
     	// received message. If not included, it uses the time the message was
     	// received on the server.
     	Time string

     	// OPTIONAL. Golang time parse string
     	TimeParse string

     	// OPTIONAL. Defaults to true. If true, the archiver will call bw2bind's "GetMetadata" on the archived URI,
     	// which inherits metadata from each of its components
     	InheritMetadata bool

     	// OPTIONAL. a list of base URIs to scan for metadata. If `<uri>` is provided, we
     	// scan `<uri>/!meta/+` for metadata keys/values
     	MetadataURIs []string

     	// OPTIONAL. a URI terminating in a metadata key that contains some kv
     	// structure of metadata, for example `/a/b/c/!meta/metadatahere`
     	MetadataBlock string

     	// OPTIONAL. a ObjectBuilder expression to search in the current message
     	// for metadata
     	MetadataExpr string
}
```

To begin archiving some set of URIs, we can create an `archive.yml` file that contains an encoding of several archive requests and
specifies where they should be placed. This is a simple, straightforward way to manage what gets archived.


```yaml
# archive.yml

# this is the prefix for all URIs in this file
Prefix: scratch.ns/abc/def
# list of archive requests. Each entry is the set of key-value pairs comprising an archive request
Archive:
	- URI: s.top/pantry/i.top/signal/cpu
	  Value: Value
 	  MetadataInherit: true
```

The `Value` expression is an "objectbuilder" expression, which is essentially a JSON-like selector (think Python dictionaries or accessing a JSON
object from Javascript) for arbitrary structs. See [https://github.com/gtfierro/ob](https://github.com/gtfierro/ob)
