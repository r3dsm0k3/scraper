package db

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/r3dsm0k3/scraper/utils"

	"github.com/dgraph-io/badger/v2"
)

type ApartmentDb struct {
	path   string
	badger *badger.DB
}

func New(path string) *ApartmentDb {
	// Open the Badger database
	// It will be created if it doesn't exist.
	badger, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		log.Fatal(err)
		return nil
	}
	db := &ApartmentDb{
		path:   path,
		badger: badger,
	}
	return db
}

func (db *ApartmentDb) Close() error {
	return db.badger.Close()
}

func normalizeKey(key string) string {
	key = strings.ToLower(key)
	return strings.ReplaceAll(key, " ", "-")
}

func (db *ApartmentDb) AddApartment(apartment *utils.PotentialApartment) error {
	err := db.badger.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(apartment)
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(normalizeKey(apartment.Location)), b)
		err = txn.SetEntry(e)
		return err
	})
	return err
}

func (db *ApartmentDb) CheckApartmentExists(location string) bool {

	err := db.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(normalizeKey(location)))
		if err != nil {
			return err
		}
		var rawApartment []byte
		err = item.Value(func(val []byte) error {
			rawApartment = append([]byte{}, val...)
			return nil
		})
		if err != nil {
			return err
		}
		// if the value exists, we try to unmarshall to the struct
		var apartmentStruct utils.PotentialApartment
		err = json.Unmarshal(rawApartment, &apartmentStruct)
		return err

	})
	if err != nil {
		return false
	}
	return true
}
