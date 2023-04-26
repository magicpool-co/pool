package dbcl

import (
	"testing"
)

func TestPrepareNamedInsert(t *testing.T) {
	tests := []struct {
		table string
		cols  []string
		query string
	}{
		{
			table: "a",
			cols:  []string{"aa"},
			query: "INSERT INTO `a`(`aa`) VALUES (:aa);",
		},
		{
			table: "a",
			cols:  []string{"aa", "bb", "cc"},
			query: "INSERT INTO `a`(`aa`, `bb`, `cc`) VALUES (:aa, :bb, :cc);",
		},
	}

	for i, tt := range tests {
		query := prepareNamedInsert(tt.table, tt.cols)
		if query != tt.query {
			t.Errorf("failed on %d: have %s, want %s", i, query, tt.query)
		}
	}
}

func TestPrepareNamedInsertUpdateWithOperator(t *testing.T) {
	tests := []struct {
		table      string
		insertCols []string
		updateCols []string
		operator   string
		query      string
	}{
		{
			table:      "a",
			insertCols: []string{"aa"},
			updateCols: []string{"aa"},
			operator:   "+",
			query:      "INSERT INTO `a`(`aa`) VALUES (:aa) ON DUPLICATE KEY UPDATE aa = aa + VALUES(aa);",
		},
		{
			table:      "a",
			insertCols: []string{"aa"},
			updateCols: []string{"aa"},
			operator:   "-",
			query:      "INSERT INTO `a`(`aa`) VALUES (:aa) ON DUPLICATE KEY UPDATE aa = aa - VALUES(aa);",
		},
		{
			table:      "a",
			insertCols: []string{"aa", "bb"},
			updateCols: []string{"aa", "bb"},
			operator:   "+",
			query:      "INSERT INTO `a`(`aa`, `bb`) VALUES (:aa, :bb) ON DUPLICATE KEY UPDATE aa = aa + VALUES(aa), bb = bb + VALUES(bb);",
		},
	}

	for i, tt := range tests {
		query := prepareNamedInsertUpdateWithOperator(tt.table, tt.insertCols, tt.updateCols, tt.operator)
		if query != tt.query {
			t.Errorf("failed on %d: have %s, want %s", i, query, tt.query)
		}
	}
}

func TestPrepareNamedInsertUpdateOverwrite(t *testing.T) {
	tests := []struct {
		table      string
		insertCols []string
		updateCols []string
		query      string
	}{
		{
			table:      "a",
			insertCols: []string{"aa"},
			updateCols: []string{"aa"},
			query:      "INSERT INTO `a`(`aa`) VALUES (:aa) ON DUPLICATE KEY UPDATE aa = VALUES(aa);",
		},
		{
			table:      "a",
			insertCols: []string{"aa", "bb"},
			updateCols: []string{"aa", "bb"},
			query:      "INSERT INTO `a`(`aa`, `bb`) VALUES (:aa, :bb) ON DUPLICATE KEY UPDATE aa = VALUES(aa), bb = VALUES(bb);",
		},
	}

	for i, tt := range tests {
		query := prepareNamedInsertUpdateOverwrite(tt.table, tt.insertCols, tt.updateCols)
		if query != tt.query {
			t.Errorf("failed on %d: have %s, want %s", i, query, tt.query)
		}
	}
}

func TestPrepareNamedUpdate(t *testing.T) {
	tests := []struct {
		table      string
		updateCols []string
		whereCols  []string
		updatedAt  bool
		query      string
	}{
		{
			table:      "a",
			updateCols: []string{"aa"},
			whereCols:  []string{"bb"},
			updatedAt:  false,
			query:      "UPDATE `a` SET aa = :aa WHERE bb = :bb;",
		},
		{
			table:      "a",
			updateCols: []string{"aa"},
			whereCols:  []string{"bb"},
			updatedAt:  true,
			query:      "UPDATE `a` SET aa = :aa, `updated_at` = CURRENT_TIMESTAMP WHERE bb = :bb;",
		},
		{
			table:      "a",
			updateCols: []string{"aa", "bb"},
			whereCols:  []string{"cc", "dd"},
			updatedAt:  true,
			query:      "UPDATE `a` SET aa = :aa, bb = :bb, `updated_at` = CURRENT_TIMESTAMP WHERE cc = :cc AND dd = :dd;",
		},
	}

	for i, tt := range tests {
		query := prepareNamedUpdate(tt.table, tt.updateCols, tt.whereCols, tt.updatedAt)
		if query != tt.query {
			t.Errorf("failed on %d: have %s, want %s", i, query, tt.query)
		}
	}
}
