package mysql

import (
	"fmt"
	"testing"
)

func TestGetTables(t *testing.T) {
	config := GetMySQLConfig()
	fmt.Printf("%v\n", config)
	client, err := NewMySQLClient(MySQLDSNParams{
		Username: config.MySQL.Username,
		Password: config.MySQL.Password,
		Hostname: config.MySQL.Hostname,
		Port:     config.MySQL.Port,
		DbName:   config.MySQL.DbName,
	})
	if err != nil {
		t.Errorf("Failed to create mysql client: %s\n", err.Error())
	}

	tables, err := client.GetTables()
	if err != nil {
		t.Errorf("Failed to get mysql tables: %s\n", err.Error())
	}
	for _, table := range tables {
		fmt.Printf("[*] Scanning %s table ...\n", table.Name)
		for _, index := range table.Indexes {
			fmt.Printf("<Index: %s, %s, %s>\n", index.IndexName, index.ColumnName, index.DataType)
			// fmt.Printf(`[%d]
			// 	- index_name  = %s\n
			// 	- column_name = %s\n
			// 	- data_type   = %s\n`,
			// 	idx, index.IndexName, index.ColumnName, index.DataType)
		}
	}
}
