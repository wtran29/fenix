package cache

import (
	"time"

	"github.com/dgraph-io/badger/v3"
)

type BadgerCache struct {
	Conn   *badger.DB
	Prefix string
}

func (bc *BadgerCache) Exists(str string) (bool, error) {
	_, err := bc.Get(str)
	if err != nil {
		return false, nil
	}
	return true, nil

}

func (bc *BadgerCache) Get(str string) (interface{}, error) {
	var fromCache []byte

	err := bc.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(str))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			fromCache = append(fromCache, val...)
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	decoded, err := decode(string(fromCache))
	if err != nil {
		return nil, err
	}

	item := decoded[str]

	return item, nil
}

func (bc *BadgerCache) Set(str string, val interface{}, expiry ...int) error {
	entry := Entry{}
	entry[str] = val
	encoded, err := encode(entry)
	if err != nil {
		return err
	}

	if len(expiry) > 0 {
		err = bc.Conn.Update(func(txn *badger.Txn) error {
			item := badger.NewEntry([]byte(str), encoded).WithTTL(time.Second * time.Duration(expiry[0]))
			err = txn.SetEntry(item)
			return err
		})
	} else {
		err = bc.Conn.Update(func(txn *badger.Txn) error {
			item := badger.NewEntry([]byte(str), encoded)
			err = txn.SetEntry(item)
			return err
		})
	}
	return nil
}

func (bc *BadgerCache) Remove(str string) error {
	err := bc.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(str))
		return err
	})
	return err
}

func (bc *BadgerCache) EmptyByMatch(str string) error {

	return bc.emptyKeysHelper(str)
}

func (bc *BadgerCache) Empty() error {

	return bc.emptyKeysHelper("")
}

func (bc *BadgerCache) emptyKeysHelper(str string) error {
	batch := bc.Conn.NewWriteBatch()
	defer batch.Cancel()

	collectSize := 100000
	keysCollected := 0

	err := bc.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.AllVersions = false
		opts.PrefetchValues = false
		iter := txn.NewIterator(opts)

		for iter.Seek([]byte(str)); iter.ValidForPrefix([]byte(str)); iter.Next() {
			k := iter.Item().KeyCopy(nil)
			batch.Delete(k)
			keysCollected++

			if keysCollected == collectSize {
				if err := batch.Flush(); err != nil {
					return err
				}
				keysCollected = 0
				batch = bc.Conn.NewWriteBatch()
				defer batch.Cancel()
			}
		}
		return nil
	})

	if err != nil {
		return err
	}
	if keysCollected > 0 {
		if err := batch.Flush(); err != nil {
			return err
		}
	}

	return nil
}
