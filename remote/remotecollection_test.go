package remote_test

import (
	"testing"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/mojlighetsministeriet/storage/collection"
	"github.com/mojlighetsministeriet/storage/remote"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestQuery(test *testing.T) {
	go func() {
		service := remote.NewService(false, false, "5M")
		service.Listen(":4528")
	}()

	time.Sleep(50 * time.Millisecond)

	remoteCollection, err := remote.NewRemoteCollection("http://localhost:4528", "test-remote-collection-authors-query")
	assert.NoError(test, err)

	name := remoteCollection.GetName()
	assert.Equal(test, "test-remote-collection-authors-query", name)

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

	for i := range authors {
		err = remoteCollection.Persist(&authors[i])
		assert.NoError(test, err)
		assert.NotEqual(test, uuid.Nil, authors[i])
	}

	authorsQueryResponse := []Author{}
	err = remoteCollection.Query(Author{Name: "Selma Lagerlöf"}, 0, &authorsQueryResponse)
	assert.NoError(test, err)
	assert.Equal(test, 1, len(authorsQueryResponse))
	assert.Equal(test, authors[0].ID, authorsQueryResponse[0].GetID())
	assert.Equal(test, authors[0].Name, authorsQueryResponse[0].Name)
	assert.Equal(test, authors[0].BirthDate, authorsQueryResponse[0].BirthDate)

	for _, author := range authors {
		err = remoteCollection.Delete(&author)
		assert.NoError(test, err)
	}

	allAuthorsEmptyResponse := []Author{}
	err = remoteCollection.LoadAll(&allAuthorsEmptyResponse, 0)
	assert.NoError(test, err)
	assert.Equal(test, 0, len(allAuthorsEmptyResponse))
}

func TestLimit(test *testing.T) {
	go func() {
		service := remote.NewService(false, false, "5M")
		service.Listen(":4529")
	}()

	time.Sleep(50 * time.Millisecond)

	remoteCollection, err := remote.NewRemoteCollection("http://localhost:4529", "test-remote-collection-authors-limit")
	assert.NoError(test, err)

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

	for i := range authors {
		err = remoteCollection.Persist(&authors[i])
		assert.NoError(test, err)
		assert.NotEqual(test, uuid.Nil, authors[i])
	}

	authorsQueryResponse := []Author{}
	err = remoteCollection.LoadAll(&authorsQueryResponse, 2)
	assert.NoError(test, err)
	assert.Equal(test, 2, len(authorsQueryResponse))

	for _, author := range authors {
		err = remoteCollection.Delete(&author)
		assert.NoError(test, err)
	}

	allAuthorsEmptyResponse := []Author{}
	err = remoteCollection.LoadAll(&allAuthorsEmptyResponse, 0)
	assert.NoError(test, err)
	assert.Equal(test, 0, len(allAuthorsEmptyResponse))
}

func TestFailDeleteWithWrongID(test *testing.T) {
	go func() {
		service := remote.NewService(false, false, "5M")
		service.Listen(":4530")
	}()

	time.Sleep(50 * time.Millisecond)

	remoteCollection, err := remote.NewRemoteCollection("http://localhost:4530", "test-remote-collection-authors-fail-delete")
	assert.NoError(test, err)

	type Author struct {
		collection.BaseEntry
		Name      string
		BirthDate time.Time
	}

	author := Author{}
	author.SetID(uuid.Must(uuid.NewV4()))
	err = remoteCollection.Delete(&author)
	assert.Error(test, err)
	assert.Equal(test, "404 Not Found (application/json; charset=utf-8): {\"message\":\"Not Found\"}", err.Error())
}
