package data

import (
	"context"
	"errors"

	"fission-basic/internal/biz"
	"fission-basic/internal/data/model"
	"fission-basic/kit/sqlx"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/samber/lo"
)

var _ biz.StudentRepo = (*Student)(nil)

type Student struct {
	data *Data
	l    *log.Helper
}

func NewStudent(d *Data, logger log.Logger) biz.StudentRepo {
	return &Student{
		data: d,
		l:    log.NewHelper(logger),
	}
}

func (s *Student) AddStudent(ctx context.Context, stu *biz.Student) error {
	stuM := model.Student{
		Name:      stu.Name,
		CreatedAt: stu.CreatedAt,
	}
	s.l.Infow("name", stu.Name)

	_, err := model.InsertStudent(ctx, s.data.db, &stuM)

	return err
}

func (s *Student) GetStudent(ctx context.Context, name string) (*biz.Student, error) {
	stuM, err := model.GetStudentByName(ctx, s.data.db, name)
	if err != nil {
		if errors.Is(err, sqlx.ErrNoRows) {
			return nil, nil
		} else {
			s.l.WithContext(ctx).Errorw("name", name, "err", err)
			return nil, err
		}
	}

	return &biz.Student{
		Name:      stuM.Name,
		CreatedAt: stuM.CreatedAt,
	}, nil
}

func (s *Student) ListStudents(ctx context.Context, offset, length uint) ([]*biz.Student, error) {
	stus, err := model.SelectStudents(ctx, s.data.db, offset, length)
	if err != nil {
		return nil, err
	}

	return lo.Map(stus,
		func(stu model.Student, _ int) *biz.Student {
			return &biz.Student{Name: stu.Name, CreatedAt: stu.CreatedAt}
		},
	), nil
}

func (s *Student) CountStudents(ctx context.Context) (int64, error) {
	return model.CountStudent(ctx, s.data.db)
}
