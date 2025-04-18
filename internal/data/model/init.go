package model

import (
	"context"
	"database/sql"
	"fission-basic/kit/sqlx"
	"fmt"
	"time"

	"github.com/didi/gendry/scanner"
)

func InitDB(ctx context.Context, db sqlx.DB) error {
	sql1 := `DROP TABLE IF EXISTS activity_info;`

	sqlx.ExecContext(ctx, db, sql1)

	sql1_1 := `CREATE TABLE activity_info (
  id varchar(256) NOT NULL COMMENT '主键',
  activity_name varchar(256) NOT NULL DEFAULT '' COMMENT '活动名称',
  activity_status varchar(54) DEFAULT 'unstart' COMMENT '活动状态：unstart：未开始；started:已开始；buffer:缓冲期；end：结束',
  created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  start_at datetime NOT NULL COMMENT '活动开始时间',
  end_at datetime NOT NULL COMMENT '结束时间（开始-结束-缓冲）',
  end_buffer_day int DEFAULT NULL COMMENT '结束时间后的缓冲天数',
  end_buffer_at datetime DEFAULT NULL COMMENT '进入缓冲期的时间',
  really_end_at datetime DEFAULT NULL COMMENT '真正结束的时间',
  cost_max double(64,2) DEFAULT NULL COMMENT '活动预算上限',
  PRIMARY KEY (id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci ROW_FORMAT=DYNAMIC COMMENT='活动表'
;`
	sqlx.ExecContext(ctx, db, sql1_1)

	sql1_1_1 := `INSERT INTO activity_info
(id, activity_name, activity_status, created_at, updated_at, start_at, end_at, end_buffer_day, end_buffer_at, really_end_at, cost_max)
VALUES('mlbb25031', 'mlbb25031', 'started', '2025-01-20 09:45:15', '2025-02-20 23:39:48', '2025-02-20 00:00:00', '2025-03-31 23:59:59', 1, NULL, NULL, 100000.0);`
	sqlx.ExecContext(ctx, db, sql1_1_1)

	sq1_1_1 := `DROP TABLE IF EXISTS feishu_report;`

	sqlx.ExecContext(ctx, db, sq1_1_1)

	sq1_1_2 := `CREATE TABLE feishu_report (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  date varchar(255) COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '日期',
  time varchar(255) COLLATE utf8mb4_general_ci DEFAULT '' COMMENT '时间',
  first_count int DEFAULT '0' COMMENT '初代数量',
  fission_count int DEFAULT '0' COMMENT '非初代数量',
  cover_count int DEFAULT '0' COMMENT '覆盖数量',
  cdk_count varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT '0,0,0,0,0,0' COMMENT '发放奖励数量',
  failed_count int DEFAULT '0' COMMENT '发送失败数量',
  timeout_count int DEFAULT '0' COMMENT '发送超时数量',
  intercept_count int DEFAULT '0' COMMENT '非白拦截数量',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='飞书监控报告'
;`
	sqlx.ExecContext(ctx, db, sq1_1_2)

	//sql2_1 := `DROP TABLE IF EXISTS help_code;`
	//sqlx.ExecContext(ctx, db, sql2_1)
	//	sql2_2 := `CREATE TABLE help_code (
	//  id int NOT NULL AUTO_INCREMENT,
	//  del tinyint(1) NOT NULL DEFAULT '0',
	//  create_time datetime DEFAULT CURRENT_TIMESTAMP,
	//  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	//  help_code varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '参团码',
	//  short_link_v0 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
	//  short_link_v1 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
	//  short_link_v2 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
	//  short_link_v3 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
	//  short_link_v4 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
	//  short_link_v5 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
	//  PRIMARY KEY (id),
	//  KEY code_idx (help_code) USING BTREE
	//) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='参团码表'
	//;`
	//	sqlx.ExecContext(ctx, db, sql2_2)

	sql3_1 := `DROP TABLE IF EXISTS official_msg_record;`
	sqlx.ExecContext(ctx, db, sql3_1)
	sql3_2 := `CREATE TABLE official_msg_record (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_id varchar(255) NOT NULL COMMENT '开团人',
  rally_code varchar(255) NOT NULL COMMENT '助力码',
  state int NOT NULL DEFAULT '1' COMMENT '状态表 1:未完成 2:已完成',
  channel varchar(255) NOT NULL DEFAULT '' COMMENT '渠道',
  language varchar(255) NOT NULL DEFAULT '' COMMENT '语言',
  generation int NOT NULL DEFAULT '0' COMMENT '代数',
  nickname varchar(255) NOT NULL DEFAULT '' COMMENT '用户昵称',
  send_time bigint NOT NULL DEFAULT '0' COMMENT '最后消息发送时间',
  PRIMARY KEY (id),
  UNIQUE KEY wa_code_uniq (wa_id,rally_code) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='官方助力消息表'
;`
	sqlx.ExecContext(ctx, db, sql3_2)

	sql4_1 := `DROP TABLE IF EXISTS receipt_msg_record;`
	sqlx.ExecContext(ctx, db, sql4_1)
	sql4_2 := `CREATE TABLE receipt_msg_record (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  msg_id varchar(255) NOT NULL COMMENT 'wa消息id',
  msg_state int NOT NULL COMMENT 'wa消息状态',
  state int NOT NULL DEFAULT '1' COMMENT '状态表 1:未完成 2:已完成',
  cost_info text COMMENT '花费实体数组json数据',
  wa_id varchar(255) DEFAULT NULL,
  pt varchar(64) NOT NULL,
  PRIMARY KEY (id,pt),
  UNIQUE KEY id_state_uniq (msg_id,msg_state,pt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='回执消息表'
PARTITION BY KEY (pt)
PARTITIONS 120 ;`
	sqlx.ExecContext(ctx, db, sql4_2)

	// 	sql5_1 := `DROP TABLE IF EXISTS student;`
	// 	sqlx.ExecContext(ctx, db, sql5_1)
	// 	sql5_2 := `CREATE TABLE student (
	//   id int NOT NULL AUTO_INCREMENT,
	//   name varchar(255) COLLATE utf8mb4_general_ci DEFAULT NULL,
	//   created_at datetime DEFAULT CURRENT_TIMESTAMP,
	//   PRIMARY KEY (id)
	// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='学生'
	// ;`
	// sqlx.ExecContext(ctx, db, sql5_2)

	sql6_1 := `DROP TABLE IF EXISTS system_config;`
	sqlx.ExecContext(ctx, db, sql6_1)
	sql6_2 := `CREATE TABLE system_config (
	 id int NOT NULL AUTO_INCREMENT,
	 del tinyint(1) NOT NULL DEFAULT '0',
	 create_time datetime DEFAULT CURRENT_TIMESTAMP,
	 update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	 param_key varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
	 param_value varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '',
	 PRIMARY KEY (id)
	) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci
	;`
	sqlx.ExecContext(ctx, db, sql6_2)

	sql7_1 := `DROP TABLE IF EXISTS unofficial_msg_record;`
	sqlx.ExecContext(ctx, db, sql7_1)
	sql7_2 := `CREATE TABLE unofficial_msg_record (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_id varchar(255) NOT NULL COMMENT '助力人',
  rally_code varchar(255) NOT NULL COMMENT '助力码',
  state int NOT NULL DEFAULT '1' COMMENT '状态表 1:未完成 2:已完成',
  channel varchar(255) NOT NULL DEFAULT '' COMMENT '渠道',
  language varchar(255) NOT NULL DEFAULT '' COMMENT '语言',
  generation int NOT NULL DEFAULT '0' COMMENT '代数',
  nickname varchar(255) NOT NULL DEFAULT '' COMMENT '用户昵称',
  send_time bigint NOT NULL DEFAULT '0' COMMENT '最后消息发送时间',
  PRIMARY KEY (id),
  UNIQUE KEY wa_code_uniq (wa_id,rally_code) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='非官方助力消息表'
;`
	sqlx.ExecContext(ctx, db, sql7_2)

	sql8_1 := `DROP TABLE IF EXISTS user_create_group;`
	sqlx.ExecContext(ctx, db, sql8_1)
	sql8_2 := `CREATE TABLE user_create_group (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  create_wa_id varchar(255) NOT NULL COMMENT '开团人ID',
  help_code varchar(255) NOT NULL COMMENT '助力码',
  generation int NOT NULL COMMENT '代次',
  create_group_time bigint NOT NULL COMMENT '开团时间',
  PRIMARY KEY (id),
  KEY wa_idx (create_wa_id) USING BTREE,
  KEY code_idx (help_code) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户开团表'
;`
	sqlx.ExecContext(ctx, db, sql8_2)

	sql9_1 := `DROP TABLE IF EXISTS user_info;`
	sqlx.ExecContext(ctx, db, sql9_1)
	sql9_2 := `CREATE TABLE user_info (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_id varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT 'wa手机号',
  help_code varchar(255) NOT NULL COMMENT '助力码',
  channel varchar(255) NOT NULL COMMENT '渠道',
  language varchar(255) NOT NULL COMMENT '语言',
  generation int NOT NULL COMMENT '代数',
  join_count int NOT NULL DEFAULT '0' COMMENT '助力人数',
  cdk_v0 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '',
  cdk_v3 varchar(255) NOT NULL DEFAULT '',
  cdk_v6 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  cdk_v9 varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL DEFAULT '',
  cdk_v12 varchar(255) NOT NULL DEFAULT '',
  cdk_v15 varchar(255) NOT NULL DEFAULT '',
  nickname varchar(255) NOT NULL COMMENT '用户昵称',
  PRIMARY KEY (id),
  UNIQUE KEY id_uniq (wa_id) USING BTREE,
  KEY idx_help_code (help_code) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户信息表'
;`
	sqlx.ExecContext(ctx, db, sql9_2)

	//sql_9_3 := `INSERT INTO user_info (id,del,create_time,update_time,wa_id,help_code,channel,language,generation,join_count,cdk_v0,cdk_v3,cdk_v6,cdk_v9,cdk_v12,cdk_v15,nickname) VALUES (1,0,'2025-02-17 17:08:36','2025-02-18 00:22:20','12345','ajds1','a','03',2,0,'222neqyxv543234gm','222tggnyrynr234gv','223g54wk3ufw234gv','2245eu2bsq6523543','4y5fc73fuqbk23554','222967w8we53234gm','测试名字');`
	//exeSql, err := sqlx.ExeSql(ctx, db, sql_9_3)
	//if err != nil {
	//	log.Errorf("insert user_info error,err:%v", err)
	//} else {
	//	log.Infof("insert user_info success,affectedRows:%v", exeSql.RowsAffected)
	//}

	sql10_1 := `DROP TABLE IF EXISTS user_join_group;`
	sqlx.ExecContext(ctx, db, sql10_1)
	sql10_2 := `CREATE TABLE user_join_group (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  join_wa_id varchar(255) NOT NULL COMMENT '助力人ID',
  help_code varchar(255) NOT NULL COMMENT '被助力码',
  join_group_time bigint NOT NULL COMMENT '助力时间',
  PRIMARY KEY (id),
  UNIQUE KEY join_uniq (join_wa_id) USING BTREE,
  KEY code_idx (help_code) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户参团表'
;`
	sqlx.ExecContext(ctx, db, sql10_2)

	sql11_1 := `DROP TABLE IF EXISTS user_remind;`
	sqlx.ExecContext(ctx, db, sql11_1)
	sql11_2 := `CREATE TABLE user_remind (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_id varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  last_send_time bigint NOT NULL COMMENT '最后消息发送时间',
  send_time_v22 bigint NOT NULL COMMENT 'v22信息发送时间',
  status_v22 int NOT NULL DEFAULT '1' COMMENT '提醒消息发送状态 1:未发送 2:已发送',
  send_time_v3 bigint NOT NULL COMMENT 'v3信息发送时间',
  status_v3 int NOT NULL DEFAULT '1' COMMENT '提醒消息发送状态 1:未发送 2:已发送',
  send_time_v36 bigint NOT NULL COMMENT 'v36信息发送时间',
  status_v36 int NOT NULL DEFAULT '1' COMMENT '提醒消息发送状态 1:未发送 2:已发送 3:不发送',
  send_time_v0 bigint NOT NULL COMMENT 'v0免费cdk信息发送时间',
  status_v0 int NOT NULL DEFAULT '1' COMMENT 'v0免费cdk提醒消息发送状态 1:未发送 2:已发送',
  PRIMARY KEY (id),
  UNIQUE KEY id_uniq (wa_id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='用户提醒表'
;`
	sqlx.ExecContext(ctx, db, sql11_2)

	//打印字符串2025_02_17 从当前日期，每次打印下一天，2.28之后是3.1
	// 解析起始日期
	startDate, _ := time.Parse("2006_01_02", "2025_02_17")

	// 打印起始日期和之后的60天
	for i := 0; i < 120; i++ { // 包括起始日期本身
		nextDate := startDate.AddDate(0, 0, i) // 添加天数
		dateStr, err := fmt.Println(nextDate.Format("2006_01_02"))
		if err != nil {
			fmt.Println("Error:", err)
		} else {
			fmt.Printf("第%d天：%s\n", dateStr, nextDate.Format("2006_01_02"))
			tableName := "wa_msg_received_" + nextDate.Format("2006_01_02")

			tableDate := fmt.Sprintf(`DROP TABLE IF EXISTS %s`, tableName)
			sqlx.ExecContext(ctx, db, tableDate)

			//生成sql语句
			sql := fmt.Sprintf(`CREATE TABLE %s (
				  id int NOT NULL AUTO_INCREMENT,
				  del tinyint(1) NOT NULL DEFAULT '0',
				  create_time datetime DEFAULT CURRENT_TIMESTAMP,
				  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
				  wa_msg_id varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'wa消息id',
				  wa_id varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
				  content text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci COMMENT '发送消息内容',
				  msg_received_time bigint NOT NULL COMMENT '消息发送时间',
				  PRIMARY KEY (id),
				  KEY msg_uniq (wa_msg_id) USING BTREE,
				  KEY user_idx (wa_id) USING BTREE
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='wa接收消息表'
				;`, tableName)
			err := sqlx.ExecContext(ctx, db, sql)
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Printf("第%d天：%s\n", dateStr, nextDate.Format("2006_01_02"))
			}
		}
	}
	sql13_1 := `DROP TABLE IF EXISTS wa_msg_remind_cost_v36;`
	sqlx.ExecContext(ctx, db, sql13_1)
	sql13_2 := `CREATE TABLE wa_msg_remind_cost_v36 (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_msg_id varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL COMMENT 'wa消息id',
  cost int NOT NULL DEFAULT '0' COMMENT '发送费用',
  PRIMARY KEY (id),
  KEY msg_idx (wa_msg_id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='wa付费续时消息费用表'
;`
	sqlx.ExecContext(ctx, db, sql13_2)

	sql14_1 := `DROP TABLE IF EXISTS wa_msg_retry;`
	sqlx.ExecContext(ctx, db, sql14_1)
	sql14_2 := `CREATE TABLE wa_msg_retry (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_id varchar(255) NOT NULL,
  wa_msg_id varchar(255) NOT NULL COMMENT 'wa消息id',
  msg_type varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT '0' COMMENT '消息类型',
  state int NOT NULL DEFAULT '1' COMMENT '发送状态 0:初始化 1:发送成功 2:发送失败 3:nx发送wa成功 4:nx发送wa失败 5:nx发送wa超时',
  content text COMMENT '发送消息内容',
  build_msg_param text COMMENT '构建消息的参数',
  send_res text COMMENT '发送牛信云返回结果',
  PRIMARY KEY (id),
  KEY msg_IDX (wa_msg_id) USING BTREE,
  KEY wa_msg_retry_state_IDX (state,wa_id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='wa重试消息表'
;`
	sqlx.ExecContext(ctx, db, sql14_2)

	sql15_1 := `DROP TABLE IF EXISTS wa_msg_send;`
	sqlx.ExecContext(ctx, db, sql15_1)
	sql15_2 := `CREATE TABLE wa_msg_send (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  pt varchar(64) NOT NULL,
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  wa_msg_id varchar(255) NOT NULL COMMENT 'wa消息id',
  wa_id varchar(255) NOT NULL,
  state int NOT NULL DEFAULT '0' COMMENT '发送状态 0:初始化 未发送；1:发送成功 2:发送失败 3:nx发送wa成功 4:nx发送wa失败 5:nx发送wa超时',
  content text COMMENT '发送消息内容',
  msg_type varchar(255) NOT NULL DEFAULT '0' COMMENT '消息类型',
  build_msg_param text COMMENT '构建消息的参数',
  send_res text COMMENT '发送牛信云返回结果',
  PRIMARY KEY (id,pt),
  KEY msg_idx (wa_msg_id) USING BTREE,
  KEY wa_msg_send_state_IDX (state,wa_id) USING BTREE
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='wa发送消息表'
PARTITION BY KEY (pt)
PARTITIONS 120
;`
	sqlx.ExecContext(ctx, db, sql15_2)

	sql16_1 := `DROP TABLE IF EXISTS email_report;`
	sqlx.ExecContext(ctx, db, sql16_1)
	sql16_2 := `
CREATE TABLE email_report (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  date varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '数据日期',
  utc int NOT NULL COMMENT '时区',
  language varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '语言',
  channel varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '渠道',
  country_code varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '国码',
  generation_count varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '代次人数',
  daily_join_count varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '每日参团人数',
  total_join_count varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL DEFAULT '' COMMENT '累计参团人数',
  count_v3 int NOT NULL DEFAULT '0' COMMENT '催团消息数',
  count_v22 int NOT NULL DEFAULT '0' COMMENT '免费续时消息数',
  count_v36 int NOT NULL DEFAULT '0' COMMENT '付费续时消息数',
  success_count int NOT NULL DEFAULT '0' COMMENT '消息发送成功数',
  failed_count int NOT NULL DEFAULT '0' COMMENT '消息发送失败数',
  timeout_count int NOT NULL DEFAULT '0' COMMENT '消息发送超时数',
  intercept_count int NOT NULL DEFAULT '0' COMMENT '非白拦截消息数',
  PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;`
	sqlx.ExecContext(ctx, db, sql16_2)

	sql17_1 := `DROP TABLE IF EXISTS upload_user_info;`
	sqlx.ExecContext(ctx, db, sql17_1)
	sql17_2 := `CREATE TABLE upload_user_info (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  phone_number varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT '要发送的号码',
  last_send_time datetime NOT NULL COMMENT '上次发送时间',
  state int NOT NULL DEFAULT '0' COMMENT '处理状态',
  PRIMARY KEY (id),
  UNIQUE KEY number_uniq (phone_number) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci COMMENT='文件上传的用户信息';`
	sqlx.ExecContext(ctx, db, sql17_2)

	sql18_1 := `DROP TABLE IF EXISTS push_event_send_message;`
	sqlx.ExecContext(ctx, db, sql18_1)
	sql18_2 := `CREATE TABLE push_event_send_message (
  id int NOT NULL AUTO_INCREMENT,
  del tinyint(1) NOT NULL DEFAULT '0',
  create_time datetime DEFAULT CURRENT_TIMESTAMP,
  update_time datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  message_id varchar(255) COLLATE utf8mb4_general_ci NOT NULL COMMENT 'wa消息id',
  cost int NOT NULL COMMENT '花费 单位0.001元',
  version int DEFAULT '0' COMMENT '引流事件版本 1 2 3 4',
  PRIMARY KEY (id),
  UNIQUE KEY msg_idx (message_id) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;`
	sqlx.ExecContext(ctx, db, sql18_2)
	return nil
}

func ExeSql(ctx context.Context, db sqlx.DB, sql string) (sql.Result, error) {
	execContext, err := db.ExecContext(ctx, sql)
	if err != nil {
		return nil, err
	}
	return execContext, nil
}

func QuerySql(ctx context.Context, db sqlx.DB, sql string) ([]map[string]interface{}, error) {
	rows, err := db.QueryxContext(ctx, sql)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	data, err := scanner.ScanMapDecode(rows)
	if err != nil {
		return nil, err
	}

	return data, nil
}
