package sessionstore

import (
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/mojlighetsministeriet/storage/collection"
	"github.com/mojlighetsministeriet/storage/remote"
	uuid "github.com/satori/go.uuid"
)

type StoredSession struct {
	collection.BaseEntry
	Data string
}

func NewStore(url string, domain string, keyPairs ...[]byte) (store *Store, err error) {
	collection, err := remote.NewRemoteCollection(url)
	if err != nil {
		return
	}

	store = &Store{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
			Domain: domain,
		},
		Collection: collection,
	}

	store.MaxAge(store.Options.MaxAge)

	return
}

type Store struct {
	Codecs     []securecookie.Codec
	Options    *sessions.Options
	Collection *remote.RemoteCollection
}

func (store *Store) MaxAge(age int) {
	store.Options.MaxAge = age

	// Set the maxAge for each securecookie instance.
	for _, codec := range store.Codecs {
		if secureCookie, ok := codec.(*securecookie.SecureCookie); ok {
			secureCookie.MaxAge(age)
		}
	}
}

func (store *Store) Get(request *http.Request, name string) (session *sessions.Session, err error) {
	return sessions.GetRegistry(request).Get(store, name)
}

func (store *Store) New(request *http.Request, name string) (session *sessions.Session, err error) {
	session = sessions.NewSession(store, name)
	options := *store.Options
	session.Options = &options
	session.IsNew = true

	if cookie, errCookie := request.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, cookie.Value, &session.Values, store.Codecs...)
		if err == nil {
			err = store.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}

	return
}

func (store *Store) Save(request *http.Request, writer http.ResponseWriter, session *sessions.Session) (err error) {
	// Delete if max-age is <= 0
	if session.Options.MaxAge <= 0 {
		err = store.erase(session)
		if err != nil {
			return
		}

		http.SetCookie(writer, sessions.NewCookie(session.Name(), "", session.Options))
		return
	}

	if session.ID == "" {
		// Because the ID is used in the filename, encode it to
		// use alphanumeric characters only.
		session.ID = uuid.Must(uuid.NewV4()).String()
	}

	err = store.save(session)
	if err != nil {
		return
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, store.Codecs...)
	if err != nil {
		return
	}

	http.SetCookie(writer, sessions.NewCookie(session.Name(), encoded, session.Options))

	return
}

func (store *Store) save(session *sessions.Session) (err error) {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, store.Codecs...)
	if err != nil {
		return
	}

	id, err := uuid.FromString(session.ID)
	if err != nil {
		return
	}

	storedSession := StoredSession{Data: encoded}
	storedSession.SetID(id)
	err = store.Collection.Persist(&storedSession)

	return
}

func (store *Store) load(session *sessions.Session) (err error) {
	storedSession := StoredSession{}

	id, err := uuid.FromString(session.ID)
	if err != nil {
		return
	}

	err = store.Collection.Load(id, &storedSession)
	if err != nil {
		return
	}

	err = securecookie.DecodeMulti(session.Name(), storedSession.Data, &session.Values, store.Codecs...)

	return
}

func (store *Store) erase(session *sessions.Session) (err error) {
	id, err := uuid.FromString(session.ID)
	if err != nil {
		return
	}

	storedSession := StoredSession{}
	storedSession.SetID(id)
	err = store.Collection.Delete(&storedSession)
	if err != nil {
		return
	}

	return
}
