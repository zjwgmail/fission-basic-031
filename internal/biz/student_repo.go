package biz

import (
	"context"
	"time"
)

type StudentRepo interface {
	AddStudent(ctx context.Context, s *Student) error

	GetStudent(ctx context.Context, name string) (*Student, error)

	ListStudents(ctx context.Context, offset, length uint) ([]*Student, error)

	CountStudents(ctx context.Context) (int64, error)
}

type Student struct {
	Name      string
	CreatedAt time.Time
}
