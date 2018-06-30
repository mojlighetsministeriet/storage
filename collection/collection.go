package collection

import (
	uuid "github.com/satori/go.uuid"
)

type EntryDoesNotExistError struct{}

func (err EntryDoesNotExistError) Error() string {
	return "Entity does not exist"
}

type EntryNotParsableError struct {
	ID             uuid.UUID
	CollectionName string
}

func (err EntryNotParsableError) Error() string {
	return "Entry " + err.CollectionName + "/" + err.ID.String() + ".json contains bad data."
}

type Entry interface {
	GetID() uuid.UUID
	SetID(uuid.UUID)
}

type BaseEntry struct {
	ID uuid.UUID
}

func (entry *BaseEntry) GetID() uuid.UUID {
	return entry.ID
}

func (entry *BaseEntry) SetID(id uuid.UUID) {
	entry.ID = id
}

type UntypedEntry map[string]interface{}

func (entry *UntypedEntry) GetID() uuid.UUID {
	id := (*entry)["ID"]
	if id == nil {
		return uuid.Nil
	}

	return uuid.Must(uuid.FromString(id.(string)))
}

func (entry *UntypedEntry) SetID(id uuid.UUID) {
	(*entry)["ID"] = id.String()
}

type CollectionsInfo []CollectionInfo

type CollectionInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Entries int    `json:"entries"`
}
