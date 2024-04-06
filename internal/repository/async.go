package repository

import (
	"context"

	"github.com/ecodeclub/ekit/sqlx"
	"github.com/tsukaychan/mercury/internal/domain"
	"github.com/tsukaychan/mercury/internal/repository/dao"
)

var ErrWaitingSMSNotFound = dao.ErrWaitingSMSNotFound

//go:generate mockgen -source=./async.go -package=repomocks -destination=mocks/async_sms_repository.mock.go AsyncSmsRepository
type AsyncSmsRepository interface {
	Insert(ctx context.Context, s domain.AsyncSms) error
	PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error)
	ReportScheduleResult(ctx context.Context, id int64, success bool) error
}

var _ AsyncSmsRepository = (*asyncSmsRepository)(nil)

type asyncSmsRepository struct {
	dao      dao.AsyncSmsDao
	RetryMax int
}

func NewAsyncSmsRepository(dao dao.AsyncSmsDao) AsyncSmsRepository {
	return &asyncSmsRepository{
		dao: dao,
	}
}

func (repo *asyncSmsRepository) Insert(ctx context.Context, as domain.AsyncSms) error {
	return repo.dao.Insert(ctx, repo.domainToEntity(as))
}

func (repo *asyncSmsRepository) PreemptWaitingSMS(ctx context.Context) (domain.AsyncSms, error) {
	as, err := repo.dao.GetWaitingSMS(ctx)
	if err != nil {
		return domain.AsyncSms{}, err
	}
	return repo.entityToDomain(as), nil
}

func (repo *asyncSmsRepository) ReportScheduleResult(ctx context.Context, id int64, success bool) error {
	if success {
		return repo.dao.MarkSuccess(ctx, id)
	}
	return repo.dao.MarkFailed(ctx, id)
}

func (repo *asyncSmsRepository) domainToEntity(sms domain.AsyncSms) dao.AsyncSms {
	return dao.AsyncSms{
		Config: sqlx.JsonColumn[dao.SMSConfig]{
			Val: dao.SMSConfig{
				TplId:  sms.TplId,
				Args:   sms.Args,
				Phones: sms.Phones,
			},
			Valid: true,
		},
		RetryMax: repo.RetryMax,
	}
}

func (repo *asyncSmsRepository) entityToDomain(sms dao.AsyncSms) domain.AsyncSms {
	return domain.AsyncSms{
		Id:       sms.Id,
		TplId:    sms.Config.Val.TplId,
		Args:     sms.Config.Val.Args,
		Phones:   sms.Config.Val.Phones,
		RetryMax: sms.RetryMax,
	}
}
