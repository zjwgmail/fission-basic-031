package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

type StudentUsecase struct {
	studentRepo StudentRepo
	l           *log.Helper
}

func NewStudentUsecase(studentRepo StudentRepo, l log.Logger) *StudentUsecase {
	return &StudentUsecase{
		studentRepo: studentRepo,
		l:           log.NewHelper(l),
	}
}

func (stu *StudentUsecase) AddStudent(ctx context.Context, name string) error {
	err := stu.studentRepo.AddStudent(ctx, &Student{
		Name:      name,
		CreatedAt: time.Now(),
	})
	if err != nil {
		return err
	}

	return nil
}

func (stu *StudentUsecase) GetStudent(ctx context.Context, name string) (*Student, error) {
	stu.l.Infow("name", name)
	return stu.studentRepo.GetStudent(ctx, name)
}

func (stu *StudentUsecase) ListStudents(ctx context.Context, offset, length uint) ([]*Student, int64, error) {
	total, err := stu.studentRepo.CountStudents(ctx)
	if err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	stus, err := stu.studentRepo.ListStudents(ctx, offset, length)
	if err != nil {
		return nil, 0, err
	}

	return stus, total, nil
}
