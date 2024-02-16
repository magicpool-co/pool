package dbcl

const (
	batchSize = 1000
)

func ExecInsert(q Querier, table string, cols []string, object interface{}) (uint64, error) {
	query := prepareNamedInsert(table, cols)
	res, err := q.NamedExec(query, object)
	if err != nil {
		return 0, err
	}

	insertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return uint64(insertID), nil
}

func ExecInsertNoID(q Querier, table string, cols []string, object interface{}) error {
	query := prepareNamedInsert(table, cols)
	_, err := q.NamedExec(query, object)

	return err
}

func execBulkInsertFromQuery(q Querier, query string, objects []interface{}) error {
	if len(objects) == 0 {
		return nil
	}

	for i := 0; i <= len(objects)/batchSize; i++ {
		startOffset := i * batchSize
		endOffset := (i + 1) * batchSize
		if endOffset >= len(objects) {
			endOffset = len(objects)
		}

		if startOffset >= endOffset {
			continue
		}

		_, err := q.NamedExec(query, objects[startOffset:endOffset])
		if err != nil {
			return err
		}
	}

	return nil
}

func ExecBulkInsert(q Querier, table string, cols []string, objects []interface{}) error {
	query := prepareNamedInsert(table, cols)

	return execBulkInsertFromQuery(q, query, objects)
}

func ExecBulkInsertUpdateAdd(
	q Querier,
	table string,
	insertCols, updateCols []string,
	objects []interface{},
) error {
	query := prepareNamedInsertUpdateWithOperator(table, insertCols, updateCols, "+")

	return execBulkInsertFromQuery(q, query, objects)
}

func ExecBulkInsertUpdateSubtract(
	q Querier,
	table string,
	insertCols, updateCols []string,
	objects []interface{},
) error {
	query := prepareNamedInsertUpdateWithOperator(table, insertCols, updateCols, "-")

	return execBulkInsertFromQuery(q, query, objects)
}

func ExecBulkInsertUpdateOverwrite(
	q Querier,
	table string,
	insertCols, updateCols []string,
	objects []interface{},
) error {
	query := prepareNamedInsertUpdateOverwrite(table, insertCols, updateCols)

	return execBulkInsertFromQuery(q, query, objects)
}

func ExecUpdate(
	q Querier,
	table string,
	updateCols, whereCols []string,
	updatedAt bool,
	obj interface{},
) error {
	query := prepareNamedUpdate(table, updateCols, whereCols, updatedAt)
	_, err := q.NamedExec(query, obj)

	return err
}
