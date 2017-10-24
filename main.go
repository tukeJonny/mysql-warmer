package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/tukejonny/mysql-warmer/mysql"
)

func MySQLWarmUp(config mysql.Config) {
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

	tables, err := client.GetTables()
	if err != nil {
		log.Fatalf("Failed to get mysql tables: %s\n", err.Error())
	}

	for _, table := range tables {
		fmt.Printf("[+] Warming up %s table...\n", table.Name)
		// インデックスが無ければ次のテーブルへ
		if len(table.Indexes) == 0 {
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

		fmt.Printf("[*] Invoke %s ...\n", stmt)
		_, err := client.Client.Query(stmt)
		if err != nil {
			log.Fatalf("Failed to exeute warmup for %s: %s", table.Name, err.Error())
		}
		fmt.Println(" done!")
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

	fmt.Println("[+] Finish")
}
