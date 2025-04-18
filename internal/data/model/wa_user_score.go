package model

import (
	"context"
	"fission-basic/kit/sqlx"
)

const tableWaUserScore = "wa_user_score_2"

type WaUserScore struct {
	Id            int    `db:"id"`
	WaId          string `db:"waid"`
	LastLoginTime string `db:"last_login_time"`
	State         int    `db:"state"`
	SocialScore   int    `db:"social_score"`
	RecurringProb string `db:"recurring_prob"`
}

func WaUserScorePageBySocialScore(ctx context.Context, db sqlx.DB, limit uint, length uint) ([]*WaUserScore, error) {
	var list []*WaUserScore
	where := map[string]interface{}{
		"_orderby": "social_score desc, id asc",
		"_limit":   []uint{limit, length},
	}
	err := sqlx.SelectContext(ctx, db, &list, tableWaUserScore, where)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func WaUserScorePageByRecurringProb(ctx context.Context, db sqlx.DB, limit uint, length uint) ([]*WaUserScore, error) {
	var list []*WaUserScore
	where := map[string]interface{}{
		"recurring_prob >": "0.3",
		"_orderby":         "recurring_prob desc, id asc",
		"_limit":           []uint{limit, length},
	}
	err := sqlx.SelectContext(ctx, db, &list, tableWaUserScore, where)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func WaUserScoreUpdateState(ctx context.Context, db sqlx.DB, waId string, state int) error {
	where := map[string]interface{}{
		"waid": waId,
	}
	return sqlx.UpdateContext(ctx, db, tableWaUserScore, where, map[string]interface{}{"state": state})
}
