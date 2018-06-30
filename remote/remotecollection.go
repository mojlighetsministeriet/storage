package remote

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/google/go-querystring/query"
	"github.com/mojlighetsministeriet/storage/collection"
	"github.com/mojlighetsministeriet/utils/httprequest"
	uuid "github.com/satori/go.uuid"
)

type responseID struct {
	ID uuid.UUID
}

func NewRemoteCollection(serviceURL string, name string) (collection *RemoteCollection, err error) {
	client, err := httprequest.NewJSONClient()
	if err != nil {
		return
	}

	collection = &RemoteCollection{
		serviceURL: purell.MustNormalizeURLString(serviceURL, purell.FlagsSafe|purell.FlagRemoveTrailingSlash),
		name:       name,
		client:     client,
	}

	return
}

type RemoteCollection struct {
	serviceURL string
	name       string
	client     *httprequest.JSONClient
}

func (collection RemoteCollection) GetName() string {
	return collection.name
}

func (collection RemoteCollection) GetURL() string {
	return collection.serviceURL + "/" + collection.GetName()
}

func (collection RemoteCollection) Persist(entry collection.Entry) (err error) {
	response := responseID{}
	err = collection.client.Post(collection.GetURL(), entry, &response)
	if err != nil {
		return
	}

	entry.SetID(response.ID)

	return
}

func (collection RemoteCollection) Delete(entry collection.Entry) (err error) {
	err = collection.client.Delete(collection.GetURL()+"/"+entry.GetID().String(), nil)
	return
}

func (collection RemoteCollection) Load(id uuid.UUID, entry collection.Entry) (err error) {
	err = collection.client.Get(collection.GetURL()+"/"+id.String(), entry)
	return
}

func (collection RemoteCollection) LoadAll(entries interface{}, limit int) (err error) {
	err = collection.client.Get(collection.GetURL()+"?limit="+strconv.Itoa(limit), &entries)
	return
}

func (collection RemoteCollection) Query(filter interface{}, limit int, entries interface{}) (err error) {
	filterValues, err := query.Values(filter)
	if err != nil {
		return
	}

	for key, values := range filterValues {
		numberOfValues := len(values)
		if (numberOfValues == 1 && values[0] == "0001-01-01T00:00:00Z") ||
			(numberOfValues == 16 && strings.Join(values, ".") == "0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0") {
			delete(filterValues, key)
		}
	}

	queryString := "limit=" + strconv.Itoa(limit) + "&" + filterValues.Encode()
	err = collection.client.Get(collection.GetURL()+"?"+queryString, &entries)
	return
}
