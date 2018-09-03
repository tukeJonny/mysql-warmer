package main

import (
	"flag"
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tukejonny/mysql-warmer/mysql"
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableColors: false,
		FullTimestamp: true,
	})
	log.SetLevel(log.DebugLevel)
}

func MySQLWarmUp(config mysql.Config) {
	log.Info("[+] Start warmup ...")
	client, err := mysql.NewMySQLClient(mysql.MySQLDSNParams{
		Username: config.MySQL.Username,
		Password: config.MySQL.Password,
		Hostname: config.MySQL.Hostname,
		Port:     config.MySQL.Port,
		DbName:   config.MySQL.DbName,
	})
	if err != nil {
		log.Fatalf("Failed to connct Mysql: %s\n", err.Error())
	}

	log.Info("[*] Get table list ...")
	tables, err := client.GetTables()
	if err != nil {
		log.Fatalf("Failed to get mysql tables: %s\n", err.Error())
	}

	for _, table := range tables {
		log.Infof("[*] Warming up %s ...", table.Name)
		// インデックスが無ければ次のテーブルへ
		if len(table.Indexes) == 0 {
			log.Warnf("[!] Skip %s", table.Name)
			continue
		}

		sumStmts, cols := func() (sumStmts []string, cols []string) {
			// 重複を取り除いたカラム名を抽出
			colMap := make(map[string]string)
			for _, index := range table.Indexes {
				colMap[index.ColumnName] = index.DataType
			}

			for col := range colMap {
				var sumStmt string
				switch colMap[col] {
				case "INT":
					sumStmt = fmt.Sprintf("SUM(`%s`)", col)
				case "DATETIME":
					sumStmt = fmt.Sprintf("SUM(UNIX_TIMESTAMP(`%s`))", col)
				case "STRING":
					sumStmt = fmt.Sprintf("SUM(LENGTH(`%s`))", col)
				}

				sumStmts = append(sumStmts, sumStmt)
				cols = append(cols, col)
			}
			return
		}()

		stmt := fmt.Sprintf(
			"SELECT %s FROM (SELECT %s FROM `%s` ORDER BY %s) as t1",
			strings.Join(sumStmts, ","),
			strings.Join(cols, ","),
			table.Name,
			strings.Join(cols, ","),
		)

		log.Info("[*] Querying ...")
		_, err := client.Client.Query(stmt)
		if err != nil {
			log.Fatalf("Failed to exeute warmup for %s: %s", table.Name, err.Error())
		}
		log.Info("[+] done!")
	}
}

func main() {
	var (
		username = flag.String("user", "root", "ユーザ名")
		password = flag.String("password", "", "パスワード")
		hostname = flag.String("host", "127.0.0.1", "MySQLホスト")
		port     = flag.Int("port", 3306, "ポート番号")
		unixsock = flag.String("sock", "", "UNIXドメインソケット")
		dbname   = flag.String("dbname", "test", "データベース名")
	)
	flag.Parse()

	config := mysql.Config{
		MySQL: mysql.MySQLConfig{
			Username: *username,
			Password: *password,
			Hostname: *hostname,
			Port:     *port,
			DbName:   *dbname,
			UnixSock: *unixsock,
		},
	}
	MySQLWarmUp(config)
}
