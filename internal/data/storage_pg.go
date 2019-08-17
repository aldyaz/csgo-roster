package data

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

// PostgresStorage is the postgres generic implementation of generic storage interface
// This is just a helper to reduce the databse boilerpolate.
// It's important for you to understand what's implemented here before you use it.
// If you don't understand, don't use it, and just implement the raw sql query :)
type PostgresStorage struct {
	db              Queryer
	tableName       string
	elemType        reflect.Type
	selectFields    string
	insertFields    string
	insertParams    string
	updateSetFields string
}

// NewPostgresStorage creates a new generic postgres storage
func NewPostgresStorage(db *sqlx.DB, tableName string, elem interface{}) *PostgresStorage {
	elemType := reflect.TypeOf(elem)
	return &PostgresStorage{
		db:              db,
		tableName:       tableName,
		elemType:        elemType,
		selectFields:    selectFields(elemType),
		insertFields:    insertFields(elemType),
		insertParams:    insertParams(elemType),
		updateSetFields: updateSetFields(elemType),
	}
}

// txFromContext returns the trasanction object from the context
func txFromContext(ctx context.Context) (Queryer, bool) {
	q, ok := ctx.Value(txKey).(Queryer)
	return q, ok
}

func selectFields(elemType reflect.Type) string {
	dbFields := []string{}
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		dbTag := field.Tag.Get("db")
		if !emptyTag(dbTag) {
			dbFields = append(dbFields, fmt.Sprintf(`"%s"`, dbTag))
		}
	}
	return strings.Join(dbFields, ", ")
}

func insertFields(elemType reflect.Type) string {
	dbFields := []string{`"createdAt"`, `"updatedAt"`}
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		dbTag := field.Tag.Get("db")
		if !readOnlyTag(dbTag) && !emptyTag(dbTag) {
			dbFields = append(dbFields, fmt.Sprintf(`"%s"`, dbTag))
		}
	}
	return strings.Join(dbFields, ", ")
}

func insertParams(elemType reflect.Type) string {
	dbParams := []string{":createdAt", ":updatedAt"}
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		dbTag := field.Tag.Get("db")
		if !readOnlyTag(dbTag) && !emptyTag(dbTag) {
			dbParams = append(dbParams, fmt.Sprintf(":%s", dbTag))
		}
	}
	return strings.Join(dbParams, ", ")
}

func updateSetFields(elemType reflect.Type) string {
	setFields := []string{`"updatedAt" = :updatedAt`}
	for i := 0; i < elemType.NumField(); i++ {
		field := elemType.Field(i)
		dbTag := field.Tag.Get("db")
		if !readOnlyTag(dbTag) && !emptyTag(dbTag) {
			setFields = append(setFields, fmt.Sprintf(`"%s" = :%s`, dbTag, dbTag))
		}
	}
	return strings.Join(setFields, ",")
}

func idTag(dbTag string) bool {
	return dbTag == "id"
}

func emptyTag(dbTag string) bool {
	emptyTags := []string{"", "-"}
	for _, t := range emptyTags {
		if dbTag == t {
			return true
		}
	}
	return false
}

func readOnlyTag(dbTag string) bool {
	readOnlyTags := []string{"id", "createdAt", "updatedAt"}
	for _, t := range readOnlyTags {
		if dbTag == t {
			return true
		}
	}
	return false
}

// Single queries an element according to the query & argument provided
func (r *PostgresStorage) Single(ctx context.Context, elem interface{}, where string, arg interface{}) error {
	db := r.db
	tx, ok := txFromContext(ctx)
	forUpdate := ""
	if ok {
		db = tx
		forUpdate = " FOR UPDATE"
	}

	statement, err := db.PrepareNamed(fmt.Sprintf(`SELECT %s FROM "%s" WHERE %s%s`,
		r.selectFields, r.tableName, where, forUpdate))
	if err != nil {
		return err
	}

	err = statement.Get(elem, arg)
	if err != nil {
		return err
	}

	return nil
}

// Where queries the elements according to the query & argument provided
func (r *PostgresStorage) Where(ctx context.Context, dest interface{}, where string, arg interface{}) error {
	db := r.db
	tx, ok := txFromContext(ctx)
	forUpdate := ""
	if ok {
		db = tx
		forUpdate = " FOR UPDATE"
	}

	statement, err := db.PrepareNamed(fmt.Sprintf(`SELECT %s FROM "%s" WHERE %s%s`,
		r.selectFields, r.tableName, where, forUpdate))
	if err != nil {
		return err
	}

	err = statement.Select(dest, arg)
	if err != nil {
		return err
	}

	return nil
}

// FindByID finds an element by its id
// it's defined in this project context that
// the id column name in the db should be "id"
func (r *PostgresStorage) FindByID(ctx context.Context, elem interface{}, id interface{}) error {
	err := r.Single(ctx, elem, `"id" = :id`, map[string]interface{}{
		"id": id,
	})
	if err != nil {
		return err
	}

	return nil
}

// FindAll finds all elements from the database.
func (r *PostgresStorage) FindAll(ctx context.Context, dest interface{}, page int, limit int) error {
	err := r.Where(ctx, dest, `true ORDER BY "id" DESC LIMIT :limit OFFSET :offset`, map[string]interface{}{
		"limit":  limit,
		"offset": (page - 1) * limit,
	})

	if err != nil {
		return err
	}

	return nil
}

// Count counts the size of elems inside database
func (r *PostgresStorage) Count(ctx context.Context) (int, error) {
	db := r.db
	tx, ok := txFromContext(ctx)
	if ok {
		db = tx
	}

	stmt, err := db.PrepareNamed(fmt.Sprintf("SELECT COUNT(*) FROM %s", r.tableName))
	if err != nil {
		return 0, err
	}

	var count int
	err = stmt.Get(&count, map[string]interface{}{})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// Insert inserts a new element into the database.
// It assumes the primary key of the table is "id" with the serial type.
// It will set the "createdAt" and "updatedAt" fields with current time.
func (r *PostgresStorage) Insert(ctx context.Context, elem interface{}) error {
	db := r.db
	tx, ok := txFromContext(ctx)
	if ok {
		db = tx
	}

	query := `INSERT INTO "%s" (%s) VALUES (%s) RETURNING %s`
	query = fmt.Sprintf(query, r.tableName, r.insertFields, r.insertParams, r.selectFields)
	statement, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}

	dbArgs := r.insertArgs(elem)
	err = statement.Get(elem, dbArgs)
	if err != nil {
		return err
	}
	return nil
}

// InsertBulk insert multiple rows at once
func (r *PostgresStorage) InsertBulk(ctx context.Context, elem interface{}) error {
	if reflect.Indirect(reflect.ValueOf(elem)).Len() == 0 {
		return nil
	}

	db := r.db
	tx, ok := txFromContext(ctx)
	if ok {
		db = tx
	}

	s := reflect.Indirect(reflect.ValueOf(elem))
	if s.Kind() != reflect.Slice {
		return errors.New("elem must be a slice")
	}

	params := []string{}
	for i := 0; i < s.Len(); i++ {
		createdAtValue := fmt.Sprintf(":%s", fmt.Sprintf("%s%d", "createdAt", i))
		updatedAtValue := fmt.Sprintf(":%s", fmt.Sprintf("%s%d", "updatedAt", i))
		columnValues := []string{createdAtValue, updatedAtValue}

		for j := 0; j < r.elemType.NumField(); j++ {
			dbTag := r.elemType.Field(j).Tag.Get("db")
			if !emptyTag(dbTag) && !readOnlyTag(dbTag) {
				columnValues = append(columnValues, fmt.Sprintf(":%s", fmt.Sprintf("%s%d", dbTag, i)))
			}
		}

		params = append(params, fmt.Sprintf("(%s)", strings.Join(columnValues, ", ")))
	}

	bindValues := map[string]interface{}{}
	for i := 0; i < s.Len(); i++ {
		createdAtIndex := fmt.Sprintf("%s%d", "createdAt", i)
		updatedAtIndex := fmt.Sprintf("%s%d", "updatedAt", i)
		bindValues[createdAtIndex] = time.Now().UTC()
		bindValues[updatedAtIndex] = time.Now().UTC()

		for j := 0; j < r.elemType.NumField(); j++ {
			dbTag := r.elemType.Field(j).Tag.Get("db")
			if !emptyTag(dbTag) && !readOnlyTag(dbTag) {
				dbTagIndex := fmt.Sprintf("%s%d", dbTag, i)
				bindValues[dbTagIndex] = reflect.Indirect(s.Index(i)).Field(j).Interface()
			}
		}
	}

	query := fmt.Sprintf(
		`INSERT INTO "%s" (%s) VALUES %s RETURNING %s`,
		r.tableName, r.insertFields, strings.Join(params, ", "), r.selectFields,
	)
	statement, err := db.PrepareNamed(query)
	if err != nil {
		return err
	}

	argsLen := s.Len()
	err = statement.Select(elem, bindValues)
	if err != nil {
		return err
	}

	// TODO
	// Currently this implementation delete n first slice
	// because sqlx select appends the result
	// find out how create slice from interface
	for i := 0; i < argsLen; i++ {
		s.Set(reflect.AppendSlice(s.Slice(0, 0), s.Slice(0+1, s.Len())))
	}

	return nil
}

func (r *PostgresStorage) insertArgs(elem interface{}) map[string]interface{} {
	res := map[string]interface{}{
		"createdAt": time.Now().UTC(),
		"updatedAt": time.Now().UTC(),
	}

	v := reflect.ValueOf(elem).Elem()
	for i := 0; i < v.NumField(); i++ {
		dbTag := r.elemType.Field(i).Tag.Get("db")
		if !readOnlyTag(dbTag) && !emptyTag(dbTag) {
			res[dbTag] = v.Field(i).Interface()
		}
	}
	return res
}

// Update updates the element in the database.
// It will update the "updatedAt" field.
func (r *PostgresStorage) Update(ctx context.Context, elem interface{}) error {
	db := r.db
	tx, ok := txFromContext(ctx)
	if ok {
		db = tx
	}
	id := r.findID(elem)
	existingElem := reflect.New(r.elemType).Interface()
	err := r.FindByID(ctx, existingElem, id)
	if err != nil {
		return err
	}

	statement, err := db.PrepareNamed(fmt.Sprintf(`
		UPDATE "%s" SET %s WHERE "id" = :id RETURNING %s`,
		r.tableName,
		r.updateSetFields,
		r.selectFields))
	if err != nil {
		return err
	}

	updateArgs := r.updateArgs(existingElem, elem)
	updateArgs["id"] = id
	err = statement.Get(elem, updateArgs)
	if err != nil {
		return err
	}
	return nil
}

// it assumes the id column named "id"
func (r *PostgresStorage) findID(elem interface{}) interface{} {
	v := reflect.ValueOf(elem).Elem()
	for i := 0; i < v.NumField(); i++ {
		dbTag := r.elemType.Field(i).Tag.Get("db")
		if idTag(dbTag) {
			return v.Field(i).Interface()
		}
	}
	return nil
}

func (r *PostgresStorage) updateArgs(existingElem interface{}, elem interface{}) map[string]interface{} {
	res := map[string]interface{}{
		"updatedAt": time.Now().UTC(),
	}

	v := reflect.ValueOf(elem).Elem()
	ev := reflect.ValueOf(existingElem).Elem()
	for i := 0; i < ev.NumField(); i++ {
		dbTag := r.elemType.Field(i).Tag.Get("db")
		if !readOnlyTag(dbTag) && !emptyTag(dbTag) {
			res[dbTag] = v.Field(i).Interface()
		}
	}
	return res
}

// Delete deletes the elem from database.
// Delete not really deletes the elem from the db, but it will set the
// "deletedAt" column to current time.
func (r *PostgresStorage) Delete(ctx context.Context, id interface{}) error {
	db := r.db
	tx, ok := txFromContext(ctx)
	if ok {
		db = tx
	}

	statement, err := db.PrepareNamed(fmt.Sprintf(`
		UPDATE "%s" SET "deletedAt" = :deletedAt WHERE "id" = :id RETURNING %s
	`, r.tableName, r.selectFields))
	if err != nil {
		return err
	}

	deleteArgs := map[string]interface{}{
		"id":        id,
		"deletedAt": time.Now().UTC(),
	}
	_, err = statement.Exec(deleteArgs)
	if err != nil {
		return err
	}
	return nil
}
