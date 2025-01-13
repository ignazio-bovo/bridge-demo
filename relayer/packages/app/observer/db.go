package observer

import (
	"go.mills.io/bitcask/v2"
)

type BitcaskDB struct {
	path string
	db   *bitcask.Bitcask
}

func NewBitcaskDB(path string) (*BitcaskDB, error) {
	db, err := bitcask.Open(path)
	if err != nil {
		return nil, err
	}
	return &BitcaskDB{path: path, db: db}, nil
}

func (d *BitcaskDB) Put(key []byte, value []byte) error {
	return d.db.Put(key, value)
}

func (d *BitcaskDB) Get(key []byte) ([]byte, error) {
	value, err := d.db.Get(key)
	return value, err
}

func (d *BitcaskDB) Delete(key []byte) error {
	return d.db.Delete(key)
}

func (d *BitcaskDB) Close() error {
	return d.db.Close()
}
