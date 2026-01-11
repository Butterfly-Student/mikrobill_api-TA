package postgres_outbound_adapter

import (
	"database/sql"

	"github.com/pkg/errors"

	outbound_port "prabogo/internal/port/outbound"
)

type adapter struct {
	db         *sql.DB
	dbexecutor outbound_port.DatabaseExecutor
}

type txWrapper struct {
	*sql.Tx
}

func (tx *txWrapper) Begin() (*sql.Tx, error) {
	return nil, errors.New("cannot start a transaction within a transaction")
}

func NewAdapter(db *sql.DB) outbound_port.DatabasePort {
	return &adapter{
		db: db,
	}
}

func (s *adapter) DoInTransaction(txFunc outbound_port.InTransaction) (out interface{}, err error) {
	var tx *sql.Tx
	reg := s
	if s.dbexecutor == nil {
		tx, err = s.db.Begin()
		if err != nil {
			return
		}
		defer func() {
			if p := recover(); p != nil {
				_ = tx.Rollback()
				switch x := p.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					// Fallback err (per specs, error strings should be lowercase w/o punctuation
					err = errors.New("unknown panic")
				}
			} else if err != nil {
				xerr := tx.Rollback() // err is non-nil; don't change it
				if xerr != nil {
					err = errors.Wrap(err, xerr.Error())
				}
			} else {
				err = tx.Commit() // err is nil; if Commit returns error update err
			}
		}()
		reg = &adapter{
			db:         s.db,
			dbexecutor: &txWrapper{Tx: tx},
		}
	}
	out, err = txFunc(reg)
	if err != nil {
		if out != nil {
			return out, err
		}

		return nil, err
	}
	return
}

func (s *adapter) Client() outbound_port.ClientDatabasePort {
	if s.dbexecutor != nil {
		return NewClientAdapter(s.dbexecutor)
	}
	return NewClientAdapter(s.db)
}

func (s *adapter) Auth() outbound_port.AuthDatabasePort {
	if s.dbexecutor != nil {
		return NewAuthAdapter(s.dbexecutor)
	}
	return NewAuthAdapter(s.db)
}

func (s *adapter) Mikrotik() outbound_port.MikrotikDatabasePort {
	if s.dbexecutor != nil {
		return NewMikrotikAdapter(s.dbexecutor)
	}
	return NewMikrotikAdapter(s.db)
}

func (s *adapter) Profile() outbound_port.ProfileDatabasePort {
	if s.dbexecutor != nil {
		return NewProfileAdapter(s.dbexecutor)
	}
	return NewProfileAdapter(s.db)
}

func (s *adapter) Customer() outbound_port.CustomerDatabasePort {
	if s.dbexecutor != nil {
		return NewCustomerAdapter(s.dbexecutor)
	}
	return NewCustomerAdapter(s.db)
}
