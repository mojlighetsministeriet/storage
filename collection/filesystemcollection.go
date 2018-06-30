package collection

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"sync"

	uuid "github.com/satori/go.uuid"
)

type FilesystemCollection struct {
	Name string
	mux  sync.Mutex
}

func (collection FilesystemCollection) createCollectionDirectory() error {
	return os.MkdirAll("collections/"+collection.GetName(), 0700)
}

func (collection FilesystemCollection) GetName() string {
	return collection.Name
}

func (collection FilesystemCollection) Persist(entry Entry) (err error) {
	collection.mux.Lock()
	defer collection.mux.Unlock()

	if entry.GetID() == uuid.Nil {
		entry.SetID(uuid.Must(uuid.NewV4()))
	}

	err = collection.createCollectionDirectory()
	if err != nil {
		return
	}

	filePath := "collections/" + collection.GetName() + "/" + entry.GetID().String() + ".json"

	serialized, err := json.Marshal(entry)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(filePath, serialized, 0600)
	return
}

func (collection FilesystemCollection) Delete(entry Entry) (err error) {
	collection.mux.Lock()
	defer collection.mux.Unlock()

	filePath := "collections/" + collection.GetName() + "/" + entry.GetID().String() + ".json"

	err = os.Remove(filePath)
	if err != nil && strings.HasSuffix(err.Error(), "no such file or directory") {
		err = EntryDoesNotExistError{}
	}

	return
}

func (collection FilesystemCollection) getDirectory() string {
	return "collections/" + collection.GetName()
}

func (collection FilesystemCollection) getFilename(id uuid.UUID) string {
	return collection.getDirectory() + "/" + id.String() + ".json"
}

func (collection FilesystemCollection) getIds() (ids []uuid.UUID, err error) {
	// TODO: Implement indexes instead and let there automatically be a ID index
	files, err := ioutil.ReadDir(collection.getDirectory())
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			err = nil
		}
	}

	if err == nil {
		for _, file := range files {
			id := uuid.Must(uuid.FromString(strings.TrimSuffix(file.Name(), ".json")))
			ids = append(ids, id)
		}
	}

	return
}

func (collection FilesystemCollection) loadRaw(id uuid.UUID) (raw []byte, err error) {
	raw, err = ioutil.ReadFile(collection.getFilename(id))
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			err = EntryDoesNotExistError{}
		}
	}
	return
}

func (collection FilesystemCollection) checkIfElementPasses(filterField reflect.Value, entryField reflect.Value) bool {
	passes := true
	filterFieldType := filterField.Type()

	if filterFieldType == reflect.TypeOf(uuid.Nil) {
		entryUUID := entryField.Interface().(uuid.UUID)
		filterUUID := filterField.Interface().(uuid.UUID)

		if filterUUID != uuid.Nil && filterUUID != entryUUID {
			passes = false
		}
	} else if filterField.Kind() == reflect.Struct {
		if !collection.passesFilter(filterField, entryField) {
			passes = false
		}
	} else if filterField.Kind() == reflect.String {
		filterString := filterField.String()
		entryString := entryField.String()

		if filterString != "" && filterString != entryString {
			passes = false
		}
	} else if filterField.Kind() == reflect.Int {
		filterInt := filterField.Int()
		entryInt := entryField.Int()

		if filterInt != 0 && filterInt != entryInt {
			passes = false
		}
	}

	return passes
}

func (collection FilesystemCollection) passesFilter(filter reflect.Value, entry reflect.Value) bool {
	passes := true

	if filter.Kind() == reflect.Struct {
		for i := 0; i < filter.NumField(); i++ {
			filterField := filter.Field(i)
			filterFieldName := filter.Type().Field(i).Name
			passes = collection.checkIfElementPasses(filterField, entry.FieldByName(filterFieldName))
			if passes == false {
				break
			}
		}
	} else if filter.Kind() == reflect.Map {
		for _, key := range filter.MapKeys() {
			filterField := filter.MapIndex(key).Elem()
			filterFieldName := key.String()

			var entryElement reflect.Value
			if entry.Kind() == reflect.Struct {
				entryElement = entry.FieldByName(filterFieldName)
			} else if entry.Kind() == reflect.Map {
				entryElement = entry.MapIndex(key).Elem()
			} else {
				passes = false
				break
			}

			passes = collection.checkIfElementPasses(filterField, entryElement)
			if passes == false {
				break
			}
		}
	} else {
		passes = false
	}

	return passes
}

func (collection FilesystemCollection) Load(id uuid.UUID, entry Entry) (err error) {
	collection.mux.Lock()
	defer collection.mux.Unlock()

	raw, err := collection.loadRaw(id)
	if err != nil {
		return
	}

	err = json.Unmarshal(raw, entry)
	if err != nil {
		return
	}

	return
}

func (collection FilesystemCollection) LoadAll(entries interface{}, limit int) (err error) {
	collection.mux.Lock()
	defer collection.mux.Unlock()

	ids, err := collection.getIds()
	if err != nil {
		return
	}

	slice := reflect.ValueOf(entries).Elem()
	elementType := slice.Type().Elem()

	added := 0
	for _, id := range ids {
		if limit != 0 && added >= limit {
			return
		}

		raw, loadError := collection.loadRaw(id)
		if loadError != nil {
			err = loadError
			return
		}

		entry := reflect.New(elementType)
		unmarshalError := json.Unmarshal(raw, entry.Interface())
		if unmarshalError != nil {
			return EntryNotParsableError{
				ID:             id,
				CollectionName: collection.GetName(),
			}
		}
		slice.Set(reflect.Append(slice, entry.Elem()))
		added++
	}

	return
}

func (collection FilesystemCollection) Query(filter interface{}, limit int, entries interface{}) (err error) {
	collection.mux.Lock()
	defer collection.mux.Unlock()

	ids, err := collection.getIds()
	if err != nil {
		return
	}

	slice := reflect.ValueOf(entries).Elem()
	elementType := slice.Type().Elem()

	appended := 0
	for _, id := range ids {
		raw, loadError := collection.loadRaw(id)
		if loadError != nil {
			err = loadError
			return
		}

		entry := reflect.New(elementType)
		unmarshalError := json.Unmarshal(raw, entry.Interface())
		if unmarshalError != nil {
			return EntryNotParsableError{
				ID:             id,
				CollectionName: collection.GetName(),
			}
		}

		entryValue := entry.Elem()

		if collection.passesFilter(reflect.ValueOf(filter), entryValue) {
			slice.Set(reflect.Append(slice, entryValue))
			appended++

			if limit != 0 && appended >= limit {
				break
			}
		}
	}

	return
}

func FilesystemCollectionsInfo() (collectionsInfo CollectionsInfo, err error) {
	collectionsInfo = CollectionsInfo{}

	files, err := ioutil.ReadDir("collections/")
	if err != nil {
		if strings.HasSuffix(err.Error(), "no such file or directory") {
			err = nil
		}
	}

	if err == nil {
		for _, file := range files {
			if file.IsDir() {
				fileCollection := FilesystemCollection{Name: file.Name()}
				ids, idsError := fileCollection.getIds()
				if idsError != nil {
					err = idsError
					return
				}

				collectionsInfo = append(collectionsInfo, CollectionInfo{
					Name:    file.Name(),
					Path:    "/" + file.Name() + "/",
					Entries: len(ids),
				})
			}
		}
	}

	return
}
