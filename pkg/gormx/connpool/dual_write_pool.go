package connpool

import (
	"context"
	"database/sql"
	"errors"
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"gorm.io/gorm"
)

var errUnknownPattern = errors.New("unknown dual write pattern")

type Pattern uint8

const (
	PatternSrcOnly Pattern = iota
	PatternSrcFirst
	PatternDstFirst
	PatternDstOnly
)

func (p Pattern) String() string {
	switch p {
	case PatternSrcOnly:
		return "src only"
	case PatternSrcFirst:
		return "src first"
	case PatternDstFirst:
		return "dst first"
	case PatternDstOnly:
		return "dst only"
	default:
		return "unknown"
	}
}

type DualWritePool struct {
	src     gorm.ConnPool
	dst     gorm.ConnPool
	pattern *atomicx.Value[Pattern]
}

func NewDualWritePool(src, dst *gorm.DB) *DualWritePool {
	return &DualWritePool{
		src:     src.ConnPool,
		dst:     dst.ConnPool,
		pattern: atomicx.NewValueOf(PatternSrcOnly),
	}
}

func (d *DualWritePool) WithPattern(pattern Pattern) {
	d.pattern.Store(pattern)
}

func (d *DualWritePool) BeginTx(ctx context.Context, opts *sql.TxOptions) (gorm.ConnPool, error) {
	pattern := d.pattern.Load()
	switch pattern {
	case PatternSrcOnly:
		return d.startOneTx(ctx, d.src, pattern, opts)
	case PatternSrcFirst:
		return d.startTwoTx(ctx, d.src, d.dst, pattern, opts)
	case PatternDstFirst:
		return d.startTwoTx(ctx, d.dst, d.src, pattern, opts)
	case PatternDstOnly:
		return d.startOneTx(ctx, d.dst, pattern, opts)
	default:
		return nil, errors.New("unknown dual write pattern")
	}
}

func (d *DualWritePool) startOneTx(ctx context.Context, conn gorm.ConnPool, pattern Pattern, opts *sql.TxOptions) (*DualWritePoolTx, error) {
	tx, err := conn.(gorm.TxBeginner).BeginTx(ctx, opts)
	return &DualWritePoolTx{
		src:     tx,
		pattern: pattern,
	}, err
}

func (d *DualWritePool) startTwoTx(ctx context.Context,
	first, second gorm.ConnPool,
	pattern Pattern,
	opts *sql.TxOptions) (*DualWritePoolTx, error) {
	src, err := first.(gorm.TxBeginner).BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	dst, err := second.(gorm.TxBeginner).BeginTx(ctx, opts)
	if err != nil {
		// logger
		_ = src.Rollback()
	}

	return &DualWritePoolTx{
		src:     src,
		dst:     dst,
		pattern: pattern,
	}, nil
}

func (d *DualWritePool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("not supported in dual write mode")
}

func (d *DualWritePool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// logger
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// logger
		}
		return res, err
	default:
		panic("unknown pattern")
	}
}

func (d *DualWritePool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("unknown pattern")
	}
}

func (d *DualWritePool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern.Load() {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic("unknown pattern")
	}
}

type DualWritePoolTx struct {
	src *sql.Tx
	dst *sql.Tx

	pattern Pattern
}

func (d *DualWritePoolTx) Commit() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Commit()
	case PatternSrcFirst:
		err := d.src.Commit()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Commit()
			if err != nil {
				// logger
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Commit()
	case PatternDstFirst:
		err := d.dst.Commit()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Commit()
			if err != nil {
				// logger
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DualWritePoolTx) Rollback() error {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.Rollback()
	case PatternSrcFirst:
		err := d.src.Rollback()
		if err != nil {
			return err
		}
		if d.dst != nil {
			err = d.dst.Rollback()
			if err != nil {
				// logger
			}
		}
		return nil
	case PatternDstOnly:
		return d.dst.Rollback()
	case PatternDstFirst:
		err := d.dst.Rollback()
		if err != nil {
			return err
		}
		if d.src != nil {
			err = d.src.Rollback()
			if err != nil {
				// logger
			}
		}
		return nil
	default:
		return errUnknownPattern
	}
}

func (d *DualWritePoolTx) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("not supported in dual write mode")
}

func (d *DualWritePoolTx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	switch d.pattern {
	case PatternSrcOnly:
		return d.src.ExecContext(ctx, query, args...)
	case PatternSrcFirst:
		res, err := d.src.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.dst == nil {
			return res, err
		}
		_, err = d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			// logger
		}
		return res, err
	case PatternDstOnly:
		return d.dst.ExecContext(ctx, query, args...)
	case PatternDstFirst:
		res, err := d.dst.ExecContext(ctx, query, args...)
		if err != nil {
			return res, err
		}
		if d.src == nil {
			return res, err
		}
		_, err = d.src.ExecContext(ctx, query, args...)
		if err != nil {
			// logger
		}
		return res, err
	default:
		panic("unknown dual write pattern")
	}
}

func (d *DualWritePoolTx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryContext(ctx, query, args...)
	default:
		panic("unknown pattern")
	}
}

func (d *DualWritePoolTx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	switch d.pattern {
	case PatternSrcOnly, PatternSrcFirst:
		return d.src.QueryRowContext(ctx, query, args...)
	case PatternDstOnly, PatternDstFirst:
		return d.dst.QueryRowContext(ctx, query, args...)
	default:
		panic("unknown pattern")
	}
}
