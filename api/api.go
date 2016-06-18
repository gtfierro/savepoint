package api

import (
	"fmt"
	messages "github.com/gtfierro/giles2/plugins/bosswave"
	"github.com/pkg/errors"
	bw "gopkg.in/immesys/bw2bind.v5"
	"strings"
)

/*
What does the API need to do?
- archive a URI with some options
- archive a set of URIs with some options
	- check for duplicates here
- remove an archive request
- remove a set of archive requests
	- optional: fail if your VK did not put it there
- scan to find archive requests
*/

// This object is a set of instructions for how to create an archivable message
// from some received PayloadObject, though really this should be able to
// operate on any object. Each ArchiveRequest acts as a translator for received
// messages into a single timeseries stream
//     type ArchiveRequest struct {
//		// AUTOPOPULATED. The entity that requested the URI to be archived.
//		FromVK string
//     	// OPTIONAL. the URI to subscribe to. Requires building a chain on the URI
//     	// from the .FromVK field. If not provided, uses the base URI of where this
//     	// ArchiveRequest was stored. For example, if this request was published
//     	// on <uri>/!meta/giles, then if the URI field was elided it would default
//     	// to <uri>.
//     	URI string
//
//     	// Extracts objects of the given Payload Object type from all messages
//     	// published on the URI. If elided, operates on all PO types.
//     	PO int
//
//     	// OPTIONAL. If provided, this is used as the stream UUID.  If not
//     	// provided, then a UUIDv3 with NAMESPACE_UUID and the URI, PO type and
//     	// Value are used.
//     	UUID string
//
//     	// expression determining how to extract the value from the received
//     	// message
//     	Value string
//
//     	// OPTIONAL. Expression determining how to extract the value from the
//     	// received message. If not included, it uses the time the message was
//     	// received on the server.
//     	Time string
//
//     	// OPTIONAL. Golang time parse string
//     	TimeParse string
//
//     	// OPTIONAL. a base URI to scan for metadata. If `<uri>` is provided, we
//     	// scan `<uri>/!meta/+` for metadata keys/values
//     	MetadataURI string
//
//     	// OPTIONAL. a URI terminating in a metadata key that contains some kv
//     	// structure of metadata, for example `/a/b/c/!meta/metadatahere`
//     	MetadataBlock string
//
//     	// OPTIONAL. a ObjectBuilder expression to search in the current message
//     	// for metadata
//     	MetadataExpr string
//     }
type ArchiveRequest messages.ArchiveRequest

// Returns true if the two ArchiveRequests are equal
func (req *ArchiveRequest) SameAs(other *ArchiveRequest) bool {
	return (req != nil && other != nil) &&
		(req.URI == other.URI) &&
		(req.PO == other.PO) &&
		(req.UUID == other.UUID) &&
		(req.Value == other.Value) &&
		(req.Time == other.Time) &&
		(req.TimeParse == other.TimeParse) &&
		(req.MetadataURI == other.MetadataURI) &&
		(req.MetadataBlock == other.MetadataBlock) &&
		(req.MetadataExpr == other.MetadataExpr)
}

// pack the object for publishing
func (req *ArchiveRequest) GetPO() (bw.PayloadObject, error) {
	return bw.CreateMsgPackPayloadObject(bw.FromDotForm("2.0.8.0"), req)
}

type API struct {
	client *bw.BW2Client
	vk     string
}

// Create a new API isntance w/ the given client and VerifyingKey.
// The verifying key is returned by any of the BW2Client.SetEntity* calls
func NewAPI(client *bw.BW2Client, vk string) *API {
	return &API{
		client: client,
		vk:     vk,
	}
}

// Attaches the archive request to the given URI. The request will be packed as a
// GilesArchiveRequestPID MsgPack object and attached to <uri>/!meta/giles.
// The URI does not have to be fully specified: if your permissions allow, you can
// also request that multiple URIs be archived using a `*` or `+` in the URI.
func (api *API) AttachArchiveRequests(uri string, requests ...*ArchiveRequest) error {
	// sanity check the parameters
	if uri == "" {
		return errors.New("Need a valid URI")
	}
	for _, request := range requests {
		if request.PO == 0 {
			return errors.New("Need a valid PO number")
		}
		if request.Value == "" {
			return errors.New("Need a Value expression")
		}
	}

	// generate the publish URI
	uriFull := strings.TrimSuffix(uri, "/") + "/!meta/giles"

	existingRequests, err := api.GetArchiveRequests(uri)
	if err != nil {
		return errors.Wrap(err, "Could not fetch existing Archive Requests")
	}
	for _, existing := range existingRequests {
		for _, request := range requests {
			if request.SameAs(existing) {
				return errors.New("Request already exists")
			}
		}
	}

	var pos []bw.PayloadObject
	for _, req := range append(existingRequests, requests...) {
		if po, err := req.GetPO(); err == nil {
			pos = append(pos, po)
		} else {
			return err
		}
	}

	fmt.Printf("ATTACHING to %s\n", uriFull)
	// attach the metadata
	err = api.client.Publish(&bw.PublishParams{
		URI:            uriFull,
		PayloadObjects: pos,
		Persist:        true,
		AutoChain:      true,
	})
	return err
}

// Returns the set of ArchiveRequest objects at all points in the URI. This method
// attaches the `/!meta/giles` suffix to the provided URI pattern, so this also works
// with `+` and `*` in the URI if your permissions allow.
func (api *API) GetArchiveRequests(uri string) ([]*ArchiveRequest, error) {
	// generate the query URI
	uri = strings.TrimSuffix(uri, "/") + "/!meta/giles"
	fmt.Printf("RETRIEVING from %s\n", uri)
	queryResults, err := api.client.Query(&bw.QueryParams{
		URI:       uri,
		AutoChain: true,
	})
	var requests []*ArchiveRequest
	if err != nil {
		return requests, err
	}
	for msg := range queryResults {
		for _, po := range msg.POs {
			if po.IsTypeDF(messages.GilesArchiveRequestPIDString) {
				var req = new(ArchiveRequest)
				if err := po.(bw.MsgPackPayloadObject).ValueInto(&req); err != nil {
					return requests, err
				}
				requests = append(requests, req)
			}
		}
	}
	return requests, nil
}

// Removes *all* ArchiveRequests at the given URI (which is extended with !meta/giles).
// If checkOwnership is true, only removes those URIs that match YOUR VK. If false, it
// removes all of them Requests, bruh. But only if your permissions allow it.
func (api *API) RemoveArchiveRequests(uri string, checkOwnership bool) error {
	uriFull := strings.TrimSuffix(uri, "/") + "/!meta/giles"
	requests, err := api.GetArchiveRequests(uri)
	if err != nil {
		return errors.Wrap(err, "Could not retrieve ArchiveRequests")
	}
	fmt.Printf("DELETING ALL on %s\n", uriFull)
	var keep []*ArchiveRequest
	for _, req := range requests {
		if checkOwnership {
			if req.FromVK != api.vk {
				keep = append(keep, req)
			}
		}
	}
	if len(keep) == 0 {
		// delete all
		return api.client.Publish(&bw.PublishParams{
			URI:            uriFull,
			PayloadObjects: []bw.PayloadObject{},
			Persist:        true,
			AutoChain:      true,
		})
	}
	err = api.AttachArchiveRequests(uri, keep...)
	if err != nil {
		return errors.Wrap(err, "Could not set ArchiveRequests")
	}
	return nil
}

// Removes only the given ArchiveRequest at the given URI. If checkOwnership is true, this will
// only remove the given ArchiveRequest if your VK placed it there. If false, it will remove
// all matching ArchiveRequests.
func (api *API) RemoveArchiveRequest(uri string, checkOwnership bool, request *ArchiveRequest) error {
	uriFull := strings.TrimSuffix(uri, "/") + "/!meta/giles"
	requests, err := api.GetArchiveRequests(uri)
	if err != nil {
		return errors.Wrap(err, "Could not retrieve ArchiveRequests")
	}
	fmt.Printf("DELETING ALL on %s\n", uriFull)
	var keep []*ArchiveRequest
	for _, req := range requests {
		if checkOwnership {
			if req.FromVK == api.vk && !request.SameAs(req) {
				keep = append(keep, req)
				continue
			}
		} else if !request.SameAs(req) {
			keep = append(keep, req)
		}
	}
	err = api.AttachArchiveRequests(uri, keep...)
	if err != nil {
		return errors.Wrap(err, "Could not set ArchiveRequests")
	}
	return nil
}
