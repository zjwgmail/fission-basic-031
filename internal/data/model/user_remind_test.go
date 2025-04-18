package model

import (
	"context"
	"testing"
)

func TestGetUserRemind(t *testing.T) {
	ctx := context.Background()
	ur, err := GetUserRemindByWaID(ctx, db, "waID")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("ur=%+v", ur)
}
