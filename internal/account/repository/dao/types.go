package dao

import (
	"context"
)

type AccountDAO interface {
	AddActivities(ctx context.Context, activities []AccountActivity) error
}
