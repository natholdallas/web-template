// Package db provides manual schema reconciliation utilities on top of GORM.
// These are separate from AutoMigrate (incremental, add-only): SyncDB brings the
// actual table structure back in line with the struct declarations (add missing
// columns, drop extra ones, reorder to match declaration order) while preserving
// data; ResetTables drops and recreates each table so the column order is exactly
// the struct declaration order.
package db

import (
	"fmt"
	"slices"
	"strings"

	"github.com/natholdallas/natools4go/orms"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// SyncDB reconciles every registered table with its struct definition. Data in
// kept columns is preserved; columns removed from the struct are dropped.
func SyncDB(tx *gorm.DB) {
	if err := syncTables(tx); err != nil {
		panic(fmt.Errorf("failed to sync database schema: %w", err))
	}
}

// ResetTables drops and recreates every registered table from its struct so the
// column order exactly matches the declaration order. All data is lost.
func ResetTables(tx *gorm.DB) {
	if err := resetTables(tx); err != nil {
		panic(fmt.Errorf("failed to reset table structures: %w", err))
	}
}

func syncTables(tx *gorm.DB) error {
	for _, model := range Models {
		if err := syncTable(tx, model); err != nil {
			return err
		}
	}
	return nil
}

func syncTable(tx *gorm.DB, model any) error {
	s, err := schemaOf(tx, model)
	if err != nil {
		return err
	}
	table := s.Table

	if !tx.Migrator().HasTable(model) {
		return orms.AutoMigrate(tx, model)
	}

	actual, err := columnPositions(tx, model)
	if err != nil {
		return err
	}
	desired := desiredColumns(s)

	// drop columns that no longer exist in the struct (skip primary key)
	for name := range actual {
		if slices.Contains(desired, name) {
			continue
		}
		field := s.LookUpField(name)
		if field != nil && field.PrimaryKey {
			continue
		}
		fmt.Printf("sync %s: drop column %s\n", table, name)
		if err := tx.Migrator().DropColumn(model, name); err != nil {
			return err
		}
	}

	// add columns that are missing from the database
	for _, name := range desired {
		if _, ok := actual[name]; ok {
			continue
		}
		fmt.Printf("sync %s: add column %s\n", table, name)
		if err := tx.Migrator().AddColumn(model, name); err != nil {
			return err
		}
	}

	// reorder columns to match the declaration order
	return reorderColumns(tx, model, s, desired)
}

func resetTables(tx *gorm.DB) error {
	switch tx.Name() {
	case "mysql", "mariadb":
		tx.Exec("SET FOREIGN_KEY_CHECKS=0")
		defer tx.Exec("SET FOREIGN_KEY_CHECKS=1")
	}

	for _, model := range Models {
		s, err := schemaOf(tx, model)
		if err != nil {
			return err
		}
		if err := tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: s.Table}).Error; err != nil {
			return err
		}
		if err := orms.AutoMigrate(tx, model); err != nil {
			return err
		}
		fmt.Printf("reset table %s\n", s.Table)
	}
	return nil
}

// schemaOf parses the GORM schema for a model value and returns it.
func schemaOf(tx *gorm.DB, model any) (*schema.Schema, error) {
	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(model); err != nil {
		return nil, err
	}
	return stmt.Schema, nil
}

// desiredColumns returns the struct-declared column names that participate in
// migration, in declaration order (embedded base models are flattened already).
func desiredColumns(s *schema.Schema) []string {
	var names []string
	for _, name := range s.DBNames {
		if field, ok := s.FieldsByDBName[name]; ok && field.IgnoreMigration {
			continue
		}
		names = append(names, name)
	}
	return names
}

// columnPositions returns the actual column names mapped to their current
// ordinal position in the database.
func columnPositions(tx *gorm.DB, model any) (map[string]int, error) {
	types, err := tx.Migrator().ColumnTypes(model)
	if err != nil {
		return nil, err
	}
	pos := map[string]int{}
	for i, col := range types {
		pos[col.Name()] = i
	}
	return pos, nil
}

// reorderColumns moves every declared column so the physical order matches the
// struct declaration. MySQL/MariaDB reorders in place; other drivers fall back
// to a rebuild that copies shared columns.
func reorderColumns(tx *gorm.DB, model any, s *schema.Schema, desired []string) error {
	switch tx.Name() {
	case "mysql", "mariadb":
		return reorderInPlace(tx, model, s, desired)
	default:
		return rebuildTable(tx, model, s, desired)
	}
}

// reorderInPlace rewrites each out-of-position column with FIRST/AFTER, which on
// MySQL also normalizes the column type/comment back to the struct declaration.
func reorderInPlace(tx *gorm.DB, model any, s *schema.Schema, desired []string) error {
	var prev string
	for i, name := range desired {
		pos, err := columnPositions(tx, model)
		if err != nil {
			return err
		}
		if _, ok := pos[name]; !ok {
			prev = name
			continue
		}
		if pos[name] == i {
			prev = name
			continue
		}
		field, _ := s.FieldsByDBName[name]
		if field == nil {
			prev = name
			continue
		}
		fullType := tx.Migrator().FullDataTypeOf(field)
		var sql string
		var values []any
		if i == 0 {
			sql = "ALTER TABLE ? MODIFY COLUMN ? ? FIRST"
			values = []any{clause.Table{Name: s.Table}, clause.Column{Name: name}, fullType}
		} else {
			sql = "ALTER TABLE ? MODIFY COLUMN ? ? AFTER ?"
			values = []any{clause.Table{Name: s.Table}, clause.Column{Name: name}, fullType, clause.Column{Name: prev}}
		}
		fmt.Printf("sync %s: reorder column %s\n", s.Table, name)
		if err := tx.Exec(sql, values...).Error; err != nil {
			return err
		}
		prev = name
	}
	return nil
}

// rebuildTable recreates a table in struct order and copies shared columns back,
// used for drivers that cannot reorder columns in place.
func rebuildTable(tx *gorm.DB, model any, s *schema.Schema, desired []string) error {
	types, err := tx.Migrator().ColumnTypes(model)
	if err != nil {
		return err
	}
	actual := map[string]bool{}
	for _, col := range types {
		actual[col.Name()] = true
	}

	tmp := s.Table + "_sync_tmp"
	if err := tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: tmp}).Error; err != nil {
		return err
	}
	if err := tx.Migrator().RenameTable(s.Table, tmp); err != nil {
		return err
	}
	if err := orms.AutoMigrate(tx, model); err != nil {
		return err
	}

	common := []string{}
	for _, name := range desired {
		if _, ok := actual[name]; ok {
			common = append(common, name)
		}
	}
	if len(common) > 0 {
		var sql strings.Builder
		sql.WriteString("INSERT INTO ? (")
		values := []any{clause.Table{Name: s.Table}}
		for i, name := range common {
			if i > 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
			values = append(values, clause.Column{Name: name})
		}
		sql.WriteString(") SELECT ")
		for i, name := range common {
			if i > 0 {
				sql.WriteString(",")
			}
			sql.WriteString("?")
			values = append(values, clause.Column{Name: name})
		}
		sql.WriteString(" FROM ?")
		values = append(values, clause.Table{Name: tmp})
		if err := tx.Exec(sql.String(), values...).Error; err != nil {
			return err
		}
	}

	if err := tx.Exec("DROP TABLE IF EXISTS ?", clause.Table{Name: tmp}).Error; err != nil {
		return err
	}
	fmt.Printf("sync %s: rebuilt columns (%d)\n", s.Table, len(common))
	return nil
}
