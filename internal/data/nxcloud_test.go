package data

import (
	"fmt"
	"testing"
	"time"

	"fission-basic/kit/sqlx"
)

var db sqlx.DB

func init() {
	d, err := sqlx.Open(&sqlx.Config{
		DriverName: "mysql",
		Server:     `root:@tcp(127.0.0.1:3306)/fission?charset=utf8mb4&parseTime=True&loc=Local`,
	})
	if err != nil {
		panic(err)
	}

	db = d
}

func TestSaveOfficialMsg(t *testing.T) {
	// data := &Data{
	// 	db: db,
	// }

	// nxCloud := NewNXCloud(data, log.NewStdLogger(os.Stdout))

	// ctx := context.TODO()
	// waID := "waID2"
	// rallyCode := "rallyCode"
	// msgID := "msgID2"
	// content := "content"
	// sendTime := time.Now().UnixMilli()

	// err := nxCloud.SaveOfficialMsg(ctx, waID, rallyCode, msgID, content, sendTime)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
}

func TestSaveUnOfficialMsg(t *testing.T) {
	// data := &Data{
	// 	db: db,
	// }

	// nxCloud := NewNXCloud(data, log.NewStdLogger(os.Stdout))

	// ctx := context.TODO()
	// waID := "un-waID"
	// rallyCode := "un-rallyCode"
	// msgID := "un-msgID2"
	// content := "content"
	// sendTime := time.Now().UnixMilli()

	// err := nxCloud.SaveUnOfficialMsg(ctx, waID, rallyCode, msgID, content, sendTime)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
}

func TestTable(t *testing.T) {
	//打印字符串2025_02_17 从当前日期，每次打印下一天，2.28之后是3.1
	// 解析起始日期
	startDate, _ := time.Parse("2006_01_02", "2025_02_17")

	// 打印起始日期和之后的60天
	for i := 0; i < 120; i++ { // 包括起始日期本身
		nextDate := startDate.AddDate(0, 0, i) // 添加天数
		//dateStr, err := fmt.Println(nextDate.Format("2006_01_02"))
		//if err != nil {
		//	fmt.Println("Error:", err)
		//} else {
		//fmt.Printf("第%d天：%s\n", dateStr, nextDate.Format("2006_01_02"))
		tableName := "wa_msg_received_" + nextDate.Format("2006_01_02")

		dropIndex := fmt.Sprintf("ALTER TABLE %s  DROP INDEX `msg_uniq`;", tableName)
		fmt.Println(dropIndex)
		addIndex := fmt.Sprintf("ALTER TABLE %s ADD INDEX `msg_idx` (`wa_msg_id`);", tableName)
		fmt.Println(addIndex)
		//}
	}
}
