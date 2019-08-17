package data

import "context"

// GenericStorage represents the generic storage
// for the domain models that matches with its database models
type GenericStorage interface {
	Single(ctx context.Context, elem interface{}, where string, arg interface{}) error
	Where(ctx context.Context, elems interface{}, where string, arg interface{}) error
	FindByID(ctx context.Context, elem interface{}, id interface{}) error
	FindAll(ctx context.Context, elems interface{}, page int, limit int) error
	Count(ctx context.Context) (int, error)
	Insert(ctx context.Context, elem interface{}) error
	InsertBulk(ctx context.Context, elem interface{}) error
	Update(ctx context.Context, elem interface{}) error
	Delete(ctx context.Context, id interface{}) error
}
