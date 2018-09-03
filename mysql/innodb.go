/// MySQL InnoDB Storage ENGINE Client
package mysql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLのカラム型種別(大きく数値型、日時型、文字列型に分ける)
var (
	INT_TYPES      = []string{"tinyint", "smallint", "mediumint", "int", "integer", "bigint", "float", "double", "real", "decimal", "numeric", "bit"}
	DATETIME_TYPES = []string{"year", "date", "time", "datetime", "timestamp"}
	STRING_TYPES   = []string{"char", "binary", "varchar", "varbinary", "tinyblob", "text", "tinytext", "mediumblob", "mediumtext", "longblob", "longtext", "enum", "set"}
)

type MySQLClient struct {
	Client *sql.DB
	DbName string
}

type MySQLIndex struct {
	IndexName  string
	ColumnName string
	DataType   string
}

type MySQLTable struct {
	Name    string
	Indexes []MySQLIndex
}

type MySQLDSNParams struct {
	Username string
	Password string
	Hostname string
	Port     int
	UnixSock string
	DbName   string
}

func NewMySQLClient(params MySQLDSNParams) (*MySQLClient, error) {
	var dsn string
	if len(params.UnixSock) > 0 {
		// Unix Domain Socketが利用可能であれば、なるべく使う
		dsn = fmt.Sprintf("%s:%s@unix(%s)/%s", params.Username, params.Password, params.UnixSock, params.DbName)
	} else {
		// Unix Domain Socketが使えないならば、TCP接続
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", params.Username, params.Password, params.Hostname, params.Port, params.DbName)
	}

	client, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &MySQLClient{
		Client: client,
		DbName: params.DbName,
	}, nil
}

func (client *MySQLClient) getIndexes(tablename string) ([]MySQLIndex, error) {
	var indexes []MySQLIndex = []MySQLIndex{}

	stmt := fmt.Sprintf(`SELECT S.index_name, S.column_name,
	(CASE WHEN C.data_type IN ('%s') THEN 'INT'
				WHEN C.data_type IN ('%s') THEN 'DATETIME'
				WHEN C.data_type IN ('%s') THEN 'STRING'
				ELSE 'UNKNOWN' END) AS data_type
	FROM information_schema.STATISTICS S INNER JOIN information_schema.COLUMNS C
		ON S.column_name = C.column_name WHERE S.table_name='%s' AND S.table_schema='%s'
	ORDER BY S.table_schema, S.table_name, S.seq_in_index`,
		strings.Join(INT_TYPES, "','"),
		strings.Join(DATETIME_TYPES, "','"),
		strings.Join(STRING_TYPES, "','"),
		tablename, client.DbName)
	rows, err := client.Client.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var indexName, columnName, dataType string
		if err = rows.Scan(&indexName, &columnName, &dataType); err != nil {
			return nil, err
		}
		indexes = append(indexes, MySQLIndex{
			IndexName:  indexName,
			ColumnName: columnName,
			DataType:   dataType,
		})
	}

	return indexes, nil
}

func (client *MySQLClient) GetTables() ([]MySQLTable, error) {
	var tables []MySQLTable

	stmt := fmt.Sprintf(`SELECT table_name FROM information_schema.tables
	WHERE table_schema='%s'
	ORDER BY table_schema, table_name`, client.DbName)
	rows, err := client.Client.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tablename string
		if err = rows.Scan(&tablename); err != nil {
			return nil, err
		}

		indexes, err := client.getIndexes(tablename)
		if err != nil {
			return nil, err
		}

		tables = append(tables, MySQLTable{
			Name:    tablename,
			Indexes: indexes,
		})
	}

	return tables, nil
}
