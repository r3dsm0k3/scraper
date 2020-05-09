package db

import (
	"encoding/json"
	"log"
	"scraper/utils"
	"strings"

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

func (db *ApartmentDb) AddApartment(apartment *utils.PotentialApartment) error {
	err := db.badger.Update(func(txn *badger.Txn) error {
		b, err := json.Marshal(apartment)
		if err != nil {
			return err
		}
		e := badger.NewEntry([]byte(strings.ToLower(apartment.Location)), b)
		err = txn.SetEntry(e)
		return err
	})
	return err
}

func (db *ApartmentDb) CheckApartmentExists(location string) bool {

	err := db.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(strings.ToLower(location)))
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
		log.Println("There was an error", err.Error())
		return false
	}
	return true
}
