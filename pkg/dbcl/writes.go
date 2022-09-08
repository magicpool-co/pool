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

func ExecBulkInsert(q Querier, table string, cols []string, objects []interface{}) error {
	if len(objects) == 0 {
		return nil
	}

	query := prepareNamedInsert(table, cols)
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

func ExecBulkInsertUpdateAdd(q Querier, table string, insertCols, updateCols []string, objects []interface{}) error {
	if len(objects) == 0 {
		return nil
	}

	query := prepareNamedInsertUpdateAdd(table, insertCols, updateCols)
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

func ExecBulkInsertUpdateOverwrite(q Querier, table string, insertCols, updateCols []string, objects []interface{}) error {
	if len(objects) == 0 {
		return nil
	}

	query := prepareNamedInsertUpdateOverwrite(table, insertCols, updateCols)
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

func ExecUpdate(q Querier, table string, updateCols, whereCols []string, updatedAt bool, obj interface{}) error {
	query := prepareNamedUpdate(table, updateCols, whereCols, updatedAt)
	_, err := q.NamedExec(query, obj)

	return err
}
