package dbcl

import (
	"fmt"
	"strings"
)

func prepareNamedInsert(table string, columns []string) string {
	names := make([]string, len(columns))
	for i, column := range columns {
		names[i] = ":" + column
		columns[i] = "`" + column + "`"
	}

	keys := strings.Join(columns, ", ")
	values := strings.Join(names, ", ")

	query := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES (%s);", table, keys, values)

	return query
}

func prepareNamedInsertUpdateWithOperator(
	table string,
	insertCols, updateCols []string,
	operator string,
) string {
	names := make([]string, len(insertCols))
	for i, col := range insertCols {
		names[i] = ":" + col
		insertCols[i] = "`" + col + "`"
	}

	keys := strings.Join(insertCols, ", ")
	values := strings.Join(names, ", ")

	updateParts := make([]string, len(updateCols))
	for i, col := range updateCols {
		updateParts[i] = fmt.Sprintf("%s = %s %s VALUES(%s)", col, col, operator, col)
	}
	update := strings.Join(updateParts, ", ")

	query := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s;",
		table, keys, values, update)

	return query
}

func prepareNamedInsertUpdateOverwrite(
	table string,
	insertCols, updateCols []string,
) string {
	names := make([]string, len(insertCols))
	for i, col := range insertCols {
		names[i] = ":" + col
		insertCols[i] = "`" + col + "`"
	}

	keys := strings.Join(insertCols, ", ")
	values := strings.Join(names, ", ")

	updateParts := make([]string, len(updateCols))
	for i, col := range updateCols {
		updateParts[i] = fmt.Sprintf("%s = VALUES(%s)", col, col)
	}
	update := strings.Join(updateParts, ", ")

	query := fmt.Sprintf("INSERT INTO `%s`(%s) VALUES (%s) ON DUPLICATE KEY UPDATE %s;",
		table, keys, values, update)

	return query
}

func prepareNamedUpdate(
	table string,
	updateCols, whereCols []string,
	updatedAt bool,
) string {
	updateParts := make([]string, len(updateCols))
	for i, column := range updateCols {
		updateParts[i] = fmt.Sprintf("%s = %s", column, ":"+column)
		updateCols[i] = "`" + column + "`"
	}

	if updatedAt {
		updateParts = append(updateParts, "`updated_at` = CURRENT_TIMESTAMP")
	}
	update := strings.Join(updateParts, ", ")

	where := ""
	if len(whereCols) > 0 {
		whereParts := make([]string, len(whereCols))
		for i, column := range whereCols {
			whereParts[i] = fmt.Sprintf("%s = %s", column, ":"+column)
			whereCols[i] = "`" + column + "`"
		}

		where = fmt.Sprintf("WHERE %s", strings.Join(whereParts, " AND "))
	}

	query := fmt.Sprintf("UPDATE `%s` SET %s %s;", table, update, where)

	return query
}
