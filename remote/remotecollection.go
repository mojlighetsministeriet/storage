package remote

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/purell"
	"github.com/google/go-querystring/query"
	"github.com/mojlighetsministeriet/storage/collection"
	"github.com/mojlighetsministeriet/utils/httprequest"
	uuid "github.com/satori/go.uuid"
)

var lastSlashPattern = regexp.MustCompile(`/([^/]+)$`)

type responseID struct {
	ID uuid.UUID
}

func NewRemoteCollection(url string) (collection *RemoteCollection, err error) {
	client, err := httprequest.NewJSONClient()
	if err != nil {
		return
	}

	url = purell.MustNormalizeURLString(url, purell.FlagsSafe|purell.FlagRemoveTrailingSlash)
	result := lastSlashPattern.FindStringSubmatch(url)
	if len(result) != 2 {
		err = errors.New("Unable to extract collection name from " + url)
		return
	}

	collection = &RemoteCollection{
		url:    url,
		name:   result[1],
		client: client,
	}

	return
}

type RemoteCollection struct {
	url    string
	name   string
	client *httprequest.JSONClient
}

func (collection RemoteCollection) GetName() string {
	return collection.name
}

func (collection RemoteCollection) Persist(entry collection.Entry) (err error) {
	response := responseID{}
	err = collection.client.Post(collection.url, entry, &response)
	if err != nil {
		return
	}

	entry.SetID(response.ID)

	return
}

func (collection RemoteCollection) Delete(entry collection.Entry) (err error) {
	err = collection.client.Delete(collection.url+"/"+entry.GetID().String(), nil)
	return
}

func (collection RemoteCollection) Load(id uuid.UUID, entry collection.Entry) (err error) {
	err = collection.client.Get(collection.url+"/"+id.String(), entry)
	return
}

func (collection RemoteCollection) LoadAll(entries interface{}, limit int) (err error) {
	err = collection.client.Get(collection.url+"?limit="+strconv.Itoa(limit), &entries)
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
	err = collection.client.Get(collection.url+"?"+queryString, &entries)
	return
}
