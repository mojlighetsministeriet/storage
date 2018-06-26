package main // import "github.com/mojlighetsministeriet/storage"

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/labstack/echo"
	"github.com/mojlighetsministeriet/storage/collection"
	"github.com/mojlighetsministeriet/utils"
	"github.com/mojlighetsministeriet/utils/server"
	uuid "github.com/satori/go.uuid"
)

var decoder = schema.NewDecoder()

func main() {
	useTLS := true
	if os.Getenv("TLS") == "disable" {
		useTLS = false
	}
	bodyLimit := utils.GetEnv("BODY_LIMIT", "5M")

	service := server.NewServer(useTLS, true, bodyLimit)

	service.GET("/", func(context echo.Context) (err error) {
		info, err := collection.FilesystemCollectionsInfo()
		if err != nil {
			return respondInternalServerError(context)
		}

		return respondOK(context, info)
	})

	service.POST("/:collection", func(context echo.Context) error {
		body, err := ioutil.ReadAll(context.Request().Body)
		if err != nil {
			return respondInternalServerError(context)
		}

		entry := collection.UntypedEntry{}
		err = json.Unmarshal(body, &entry)
		if err != nil {
			return respondStringBadRequest(context, "Invalid JSON")
		}

		entryCollection := collection.FilesystemCollection{Name: context.Param("collection")}
		err = entryCollection.Persist(&entry)
		if err == nil {
			return respondOK(context, struct {
				ID uuid.UUID
			}{
				entry.GetID(),
			})
		}

		return respondInternalServerError(context)
	})

	service.GET("/:collection", func(context echo.Context) (err error) {
		limit := 0
		limitString := context.QueryParam("limit")
		if limitString != "" {
			limit, err = strconv.Atoi(limitString)
			if err != nil {
				return respondStringBadRequest(context, "Limit must be integer")
			}
		}

		filter := make(map[string]interface{})
		for key, value := range context.QueryParams() {
			filter[key] = value[0]
		}

		delete(filter, "limit")

		entryCollection := collection.FilesystemCollection{Name: context.Param("collection")}
		entries := []collection.UntypedEntry{}

		if len(filter) > 0 {
			err = entryCollection.Query(filter, limit, &entries)
		} else {
			err = entryCollection.LoadAll(&entries, limit)
		}

		if err != nil {
			return respondInternalServerError(context)
		}

		return respondOK(context, entries)
	})

	service.GET("/:collection/:id", func(context echo.Context) error {
		id, err := uuid.FromString(context.Param("id"))
		if err != nil {
			return respondStringBadRequest(context, "Invalid UUID")
		}

		entryCollection := collection.FilesystemCollection{Name: context.Param("collection")}
		entry := collection.UntypedEntry{}
		err = entryCollection.Load(id, &entry)
		if err != nil {
			return respondNotFound(context)
		}

		return respondOK(context, entry)
	})

	service.DELETE("/:collection/:id", func(context echo.Context) error {
		id, err := uuid.FromString(context.Param("id"))
		if err != nil {
			return respondStringBadRequest(context, "Invalid UUID")
		}

		entryCollection := collection.FilesystemCollection{Name: context.Param("collection")}
		entry := collection.UntypedEntry{}
		entry.SetID(id)
		err = entryCollection.Delete(&entry)
		if err != nil {
			if err.Error() == (collection.EntryDoesNotExistError{}).Error() {
				return respondNotFound(context)
			}

			return respondInternalServerError(context)
		}

		return respondEmptyOK(context)
	})

	service.Listen(":" + utils.GetEnv("PORT", "443"))
}

func respondStringBadRequest(context echo.Context, message string) error {
	return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\""+message+"\"}"))
}

func respondEmptyBadRequest(context echo.Context) error {
	return context.JSONBlob(http.StatusBadRequest, []byte("{\"message\":\"Bad Request\"}"))
}

func respondNotFound(context echo.Context) error {
	return context.JSONBlob(http.StatusNotFound, []byte("{\"message\":\"Not Found\"}"))
}

func respondInternalServerError(context echo.Context) error {
	return context.JSONBlob(http.StatusInternalServerError, []byte("{\"message\":\"Internal Server Error\"}"))
}

func respondOK(context echo.Context, data interface{}) error {
	return context.JSON(http.StatusOK, data)
}

func respondEmptyOK(context echo.Context) error {
	return context.JSONBlob(http.StatusOK, []byte("{\"message\":\"OK\"}"))
}
