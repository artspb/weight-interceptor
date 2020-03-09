package storage

import (
	"github.com/boltdb/bolt"
	"log"
)

const (
	dbName     = "DB"
	dbRequest  = "REQUEST"
	dbResponse = "RESPONSE"
)

var db *bolt.DB

func init() {
	var err error
	db, err = bolt.Open("../storage.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte(dbName))
		if err != nil {
			log.Fatal(err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(dbRequest))
		if err != nil {
			log.Fatal(err)
		}
		_, err = root.CreateBucketIfNotExists([]byte(dbResponse))
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func Request(timestamp string, request []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbName)).Bucket([]byte(dbRequest))
		return bucket.Put([]byte(timestamp), request)
	})
}

func Response(timestamp string, response []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(dbName)).Bucket([]byte(dbResponse))
		return bucket.Put([]byte(timestamp), response)
	})
}

func Shutdown() error {
	return db.Close()
}
