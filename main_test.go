package main

import (
	"os"
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/storage/collection"
	"github.com/mojlighetsministeriet/utils/httprequest"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(test *testing.T) {
	os.Setenv("PORT", "3526")
	os.Setenv("TLS", "disable")

	go func() {
		main()
	}()

	time.Sleep(50 * time.Millisecond)

	client, err := httprequest.NewClient()
	assert.NoError(test, err)

	index, err := client.Get("http://localhost:3526/")
	assert.NoError(test, err)
	assert.Equal(test, true, len(string(index)) > 1)
}

func TestPostGetDelete(test *testing.T) {
	os.Setenv("PORT", "3527")
	os.Setenv("TLS", "disable")

	go func() {
		main()
	}()

	time.Sleep(50 * time.Millisecond)

	type Author struct {
		collection.BaseEntry
		Name      string
		BirthDate time.Time
	}

	birthDate, _ := time.Parse(time.RFC3339, "1858-10-20T00:00:00Z")
	author := Author{
		Name:      "Selma Lagerlöf",
		BirthDate: birthDate,
	}

	client, err := httprequest.NewJSONClient()
	assert.NoError(test, err)

	type ResponseID struct {
		ID uuid.UUID
	}
	url := "http://localhost:3527/test-authors"
	idResponse := ResponseID{}
	err = client.Post(url, &author, &idResponse)
	assert.NoError(test, err)
	assert.NotEqual(test, uuid.Nil, idResponse.ID)

	url = "http://localhost:3527/test-authors/"
	getAllResponse := []Author{}
	err = client.Get(url, &getAllResponse)
	assert.NoError(test, err)
	assert.Equal(test, 1, len(getAllResponse))
	assert.Equal(test, idResponse.ID, getAllResponse[0].GetID())
	assert.Equal(test, author.Name, getAllResponse[0].Name)
	assert.Equal(test, author.BirthDate, getAllResponse[0].BirthDate)

	url = "http://localhost:3527/test-authors/" + idResponse.ID.String()
	getResponse := Author{}
	err = client.Get(url, &getResponse)
	assert.NoError(test, err)
	assert.Equal(test, idResponse.ID, getResponse.GetID())
	assert.Equal(test, author.Name, getResponse.Name)
	assert.Equal(test, author.BirthDate, getResponse.BirthDate)

	url = "http://localhost:3527/test-authors/" + idResponse.ID.String()
	err = client.Delete(url, nil)
	assert.NoError(test, err)

	url = "http://localhost:3527/test-authors/" + idResponse.ID.String()
	err = client.Get(url, nil)
	assert.Error(test, err)
	assert.Equal(test, "404 Not Found (application/json; charset=utf-8): {\"message\":\"Not Found\"}", err.Error())
}

func TestQuery(test *testing.T) {
	os.Setenv("PORT", "3528")
	os.Setenv("TLS", "disable")

	go func() {
		main()
	}()

	time.Sleep(50 * time.Millisecond)

	type Author struct {
		collection.BaseEntry
		Name      string
		BirthDate time.Time
	}

	birthDate, _ := time.Parse(time.RFC3339, "1858-10-20T00:00:00Z")
	authors := []Author{
		Author{
			Name:      "Selma Lagerlöf",
			BirthDate: birthDate,
		},
		Author{
			Name:      "Anna Andersson",
			BirthDate: birthDate,
		},
		Author{
			Name:      "Lena Hansson",
			BirthDate: birthDate,
		},
	}

	client, err := httprequest.NewJSONClient()
	assert.NoError(test, err)

	type ResponseID struct {
		ID uuid.UUID
	}
	url := "http://localhost:3528/test-authors"

	for i := range authors {
		idResponse := ResponseID{}
		err = client.Post(url, &authors[i], &idResponse)
		assert.NoError(test, err)
		assert.NotEqual(test, uuid.Nil, idResponse.ID)
		authors[i].SetID(idResponse.ID)
	}

	url = "http://localhost:3528/test-authors/?Name=Selma+Lagerlöf"
	authorsQueryResponse := []Author{}
	err = client.Get(url, &authorsQueryResponse)
	assert.NoError(test, err)
	assert.Equal(test, 1, len(authorsQueryResponse))
	assert.Equal(test, authors[0].ID, authorsQueryResponse[0].GetID())
	assert.Equal(test, authors[0].Name, authorsQueryResponse[0].Name)
	assert.Equal(test, authors[0].BirthDate, authorsQueryResponse[0].BirthDate)

	for _, author := range authors {
		url = "http://localhost:3528/test-authors/" + author.GetID().String()
		err = client.Delete(url, nil)
		assert.NoError(test, err)
	}

	url = "http://localhost:3528/test-authors/"
	allAuthorsEmptyResponse := []Author{}
	err = client.Get(url, &allAuthorsEmptyResponse)
	assert.NoError(test, err)
	assert.Equal(test, 0, len(allAuthorsEmptyResponse))
}

func TestLimit(test *testing.T) {
	os.Setenv("PORT", "3529")
	os.Setenv("TLS", "disable")

	go func() {
		main()
	}()

	time.Sleep(50 * time.Millisecond)

	type Author struct {
		collection.BaseEntry
		Name      string
		BirthDate time.Time
	}

	birthDate, _ := time.Parse(time.RFC3339, "1858-10-20T00:00:00Z")
	authors := []Author{
		Author{
			Name:      "Selma Lagerlöf",
			BirthDate: birthDate,
		},
		Author{
			Name:      "Anna Andersson",
			BirthDate: birthDate,
		},
		Author{
			Name:      "Lena Hansson",
			BirthDate: birthDate,
		},
	}

	client, err := httprequest.NewJSONClient()
	assert.NoError(test, err)

	type ResponseID struct {
		ID uuid.UUID
	}
	url := "http://localhost:3529/test-authors-limit"

	for i := range authors {
		idResponse := ResponseID{}
		err = client.Post(url, &authors[i], &idResponse)
		assert.NoError(test, err)
		assert.NotEqual(test, uuid.Nil, idResponse.ID)
		authors[i].SetID(idResponse.ID)
	}

	url = "http://localhost:3529/test-authors-limit/?limit=2"
	authorsQueryResponse := []Author{}
	err = client.Get(url, &authorsQueryResponse)
	assert.NoError(test, err)
	assert.Equal(test, 2, len(authorsQueryResponse))

	for _, author := range authors {
		url = "http://localhost:3529/test-authors-limit/" + author.GetID().String()
		err = client.Delete(url, nil)
		assert.NoError(test, err)
	}

	url = "http://localhost:3529/test-authors-limit/"
	allAuthorsEmptyResponse := []Author{}
	err = client.Get(url, &allAuthorsEmptyResponse)
	assert.NoError(test, err)
	assert.Equal(test, 0, len(allAuthorsEmptyResponse))
}

func TestFailDeleteLimit(test *testing.T) {
	os.Setenv("PORT", "3530")
	os.Setenv("TLS", "disable")

	go func() {
		main()
	}()

	time.Sleep(50 * time.Millisecond)

	client, err := httprequest.NewJSONClient()
	assert.NoError(test, err)

	url := "http://localhost:3530/test-authors/" + uuid.Must(uuid.NewV4()).String()
	err = client.Delete(url, nil)
	assert.Error(test, err)
	assert.Equal(test, "404 Not Found (application/json; charset=utf-8): {\"message\":\"Not Found\"}", err.Error())
}
