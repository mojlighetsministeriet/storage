package collection_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mojlighetsministeriet/storage/collection"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

type Author struct {
	collection.BaseEntry
	Name      string
	BirthDate time.Time
}

type Book struct {
	collection.BaseEntry
	Title  string
	ISBN   string
	Author uuid.UUID
	Rating int
}

func TestPersistLoadDelete(test *testing.T) {
	authors := collection.FilesystemCollection{Name: "authors"}

	birthDate, _ := time.Parse(time.RFC3339, "1858-10-20T00:00:00Z")
	author := Author{
		Name:      "Selma Lagerlöf",
		BirthDate: birthDate,
	}
	defer authors.Delete(&author)

	err := authors.Persist(&author)
	assert.NoError(test, err)

	authorFound := Author{}
	err = authors.Load(author.GetID(), &authorFound)
	assert.NoError(test, err)
	assert.NotEqual(test, nil, authorFound)
	assert.NotEqual(test, uuid.Nil, authorFound.GetID())
	assert.Equal(test, author.GetID(), authorFound.GetID())
	assert.Equal(test, author.Name, authorFound.Name)
	assert.Equal(test, author.BirthDate, authorFound.BirthDate)

	err = authors.Delete(&author)
	assert.NoError(test, err)
}

func TestShouldFailDeletingMissingEntry(test *testing.T) {
	authors := collection.FilesystemCollection{Name: "authors"}

	author := Author{}
	author.SetID(uuid.Must(uuid.NewV4()))
	err := authors.Delete(&author)
	assert.Error(test, err)
	assert.Equal(test, "Entity does not exist", err.Error())
}

func TestShouldFailLoadingMissingEntry(test *testing.T) {
	authors := collection.FilesystemCollection{Name: "authors"}
	author := Author{}
	err := authors.Load(uuid.Must(uuid.NewV4()), &author)
	assert.Error(test, err)
	assert.Equal(test, "Entity does not exist", err.Error())
}

func TestLoadAll(test *testing.T) {
	books := collection.FilesystemCollection{Name: "books"}

	nilsHolgersson := Book{
		Title: "Nils Holgerssons underbara resa genom Sverige",
		ISBN:  "9789176631874",
	}
	gostaBerlingsSaga := Book{
		Title: "Gösta Berlings saga",
		ISBN:  "9789174296051",
	}
	enHerrgardssagen := Book{
		Title: "En herrgårdssägen",
		ISBN:  "9789174296150",
	}
	defer books.Delete(&nilsHolgersson)
	defer books.Delete(&gostaBerlingsSaga)
	defer books.Delete(&enHerrgardssagen)

	err := books.Persist(&nilsHolgersson)
	assert.NoError(test, err)
	err = books.Persist(&gostaBerlingsSaga)
	assert.NoError(test, err)
	err = books.Persist(&enHerrgardssagen)
	assert.NoError(test, err)

	booksFound := []Book{}
	err = books.LoadAll(&booksFound, 2)
	assert.NoError(test, err)
	assert.Equal(test, 2, len(booksFound))

	expectedTitles := []string{nilsHolgersson.Title, gostaBerlingsSaga.Title, enHerrgardssagen.Title}
	for _, book := range booksFound {
		assert.Contains(test, expectedTitles, book.Title)
	}
}

func TestQuery(test *testing.T) {
	books := collection.FilesystemCollection{Name: "books"}

	nilsHolgersson := Book{
		Title:  "Nils Holgerssons underbara resa genom Sverige",
		ISBN:   "9789176631874",
		Author: uuid.Must(uuid.FromString("be4346f2-0721-45d0-b52f-218714aae7a8")),
		Rating: 2,
	}
	gostaBerlingsSaga := Book{
		Title:  "Gösta Berlings saga",
		ISBN:   "9789174296051",
		Rating: 2,
	}
	enHerrgardssagen := Book{
		Title:  "En herrgårdssägen",
		ISBN:   "9789174296150",
		Author: uuid.Must(uuid.FromString("be4346f2-0721-45d0-b52f-218714aae7a8")),
		Rating: 3,
	}

	defer books.Delete(&nilsHolgersson)
	defer books.Delete(&gostaBerlingsSaga)
	defer books.Delete(&enHerrgardssagen)

	err := books.Persist(&nilsHolgersson)
	assert.NoError(test, err)
	err = books.Persist(&gostaBerlingsSaga)
	assert.NoError(test, err)
	err = books.Persist(&enHerrgardssagen)
	assert.NoError(test, err)

	booksFound := []Book{}
	err = books.Query(Book{
		BaseEntry: enHerrgardssagen.BaseEntry,
		Title:     enHerrgardssagen.Title,
		Author:    enHerrgardssagen.Author,
	}, 0, &booksFound)

	assert.NoError(test, err)
	assert.Equal(test, 1, len(booksFound))
	assert.Equal(test, enHerrgardssagen.Title, booksFound[0].Title)

	booksFound = []Book{}
	err = books.Query(Book{Title: gostaBerlingsSaga.Title}, 0, &booksFound)

	assert.NoError(test, err)
	assert.Equal(test, 1, len(booksFound))
	assert.Equal(test, gostaBerlingsSaga.Title, booksFound[0].Title)

	booksFound = []Book{}
	err = books.Query(Book{Rating: 2}, 0, &booksFound)

	assert.NoError(test, err)
	assert.Equal(test, 2, len(booksFound))

	booksFound = []Book{}
	filter := make(map[string]interface{})
	filter["Rating"] = enHerrgardssagen.Rating
	err = books.Query(filter, 0, &booksFound)

	assert.NoError(test, err)
	assert.Equal(test, 1, len(booksFound))
	assert.Equal(test, enHerrgardssagen.Title, booksFound[0].Title)

	booksFound = []Book{}
	filter = make(map[string]interface{})
	filter["Title"] = enHerrgardssagen.Title
	err = books.Query(filter, 0, &booksFound)

	assert.NoError(test, err)
	assert.Equal(test, 1, len(booksFound))
	assert.Equal(test, enHerrgardssagen.Title, booksFound[0].Title)
}

func TestUntypedEntry(test *testing.T) {
	data := collection.UntypedEntry{}
	body := []byte(`{"Title":"Pippi Långstrump","ISBN":"9789129703771"}`)
	err := json.Unmarshal(body, &data)

	id := uuid.Must(uuid.NewV4())
	assert.Equal(test, uuid.Nil.String(), data.GetID().String())
	data.SetID(id)

	assert.NoError(test, err)
	assert.Equal(test, id.String(), data.GetID().String())
	assert.Equal(test, "Pippi Långstrump", data["Title"])
	assert.Equal(test, "9789129703771", data["ISBN"])
}

func TestFilesystemCollectionsInfo(test *testing.T) {
	sections := collection.FilesystemCollection{Name: "sections"}

	type Section struct {
		collection.BaseEntry
		Name string
	}

	section := Section{
		Name: "Control room",
	}

	defer sections.Delete(&section)

	err := sections.Persist(&section)
	assert.NoError(test, err)

	info, err := collection.FilesystemCollectionsInfo()
	assert.NoError(test, err)
	assert.Contains(test, info, collection.CollectionInfo{Name: "sections", Path: "/sections/", Entries: 1})
}
