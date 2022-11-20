// Code generated by SQLBoiler 4.13.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package sqlboilerPSQL

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/strmangle"
)

// DishOccurrence is an object representing the database table.
type DishOccurrence struct {
	ID     int       `boil:"id" json:"id" toml:"id" yaml:"id"`
	DishID int       `boil:"dish_id" json:"dish_id" toml:"dish_id" yaml:"dish_id"`
	Date   time.Time `boil:"date" json:"date" toml:"date" yaml:"date"`

	R *dishOccurrenceR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L dishOccurrenceL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var DishOccurrenceColumns = struct {
	ID     string
	DishID string
	Date   string
}{
	ID:     "id",
	DishID: "dish_id",
	Date:   "date",
}

var DishOccurrenceTableColumns = struct {
	ID     string
	DishID string
	Date   string
}{
	ID:     "dish_occurrences.id",
	DishID: "dish_occurrences.dish_id",
	Date:   "dish_occurrences.date",
}

// Generated where

type whereHelperint struct{ field string }

func (w whereHelperint) EQ(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.EQ, x) }
func (w whereHelperint) NEQ(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.NEQ, x) }
func (w whereHelperint) LT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.LT, x) }
func (w whereHelperint) LTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.LTE, x) }
func (w whereHelperint) GT(x int) qm.QueryMod  { return qmhelper.Where(w.field, qmhelper.GT, x) }
func (w whereHelperint) GTE(x int) qm.QueryMod { return qmhelper.Where(w.field, qmhelper.GTE, x) }
func (w whereHelperint) IN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereIn(fmt.Sprintf("%s IN ?", w.field), values...)
}
func (w whereHelperint) NIN(slice []int) qm.QueryMod {
	values := make([]interface{}, 0, len(slice))
	for _, value := range slice {
		values = append(values, value)
	}
	return qm.WhereNotIn(fmt.Sprintf("%s NOT IN ?", w.field), values...)
}

type whereHelpertime_Time struct{ field string }

func (w whereHelpertime_Time) EQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.EQ, x)
}
func (w whereHelpertime_Time) NEQ(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.NEQ, x)
}
func (w whereHelpertime_Time) LT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LT, x)
}
func (w whereHelpertime_Time) LTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.LTE, x)
}
func (w whereHelpertime_Time) GT(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GT, x)
}
func (w whereHelpertime_Time) GTE(x time.Time) qm.QueryMod {
	return qmhelper.Where(w.field, qmhelper.GTE, x)
}

var DishOccurrenceWhere = struct {
	ID     whereHelperint
	DishID whereHelperint
	Date   whereHelpertime_Time
}{
	ID:     whereHelperint{field: "\"dish_occurrences\".\"id\""},
	DishID: whereHelperint{field: "\"dish_occurrences\".\"dish_id\""},
	Date:   whereHelpertime_Time{field: "\"dish_occurrences\".\"date\""},
}

// DishOccurrenceRels is where relationship names are stored.
var DishOccurrenceRels = struct {
	Dish string
}{
	Dish: "Dish",
}

// dishOccurrenceR is where relationships are stored.
type dishOccurrenceR struct {
	Dish *Dish `boil:"Dish" json:"Dish" toml:"Dish" yaml:"Dish"`
}

// NewStruct creates a new relationship struct
func (*dishOccurrenceR) NewStruct() *dishOccurrenceR {
	return &dishOccurrenceR{}
}

func (r *dishOccurrenceR) GetDish() *Dish {
	if r == nil {
		return nil
	}
	return r.Dish
}

// dishOccurrenceL is where Load methods for each relationship are stored.
type dishOccurrenceL struct{}

var (
	dishOccurrenceAllColumns            = []string{"id", "dish_id", "date"}
	dishOccurrenceColumnsWithoutDefault = []string{"dish_id", "date"}
	dishOccurrenceColumnsWithDefault    = []string{"id"}
	dishOccurrencePrimaryKeyColumns     = []string{"id"}
	dishOccurrenceGeneratedColumns      = []string{}
)

type (
	// DishOccurrenceSlice is an alias for a slice of pointers to DishOccurrence.
	// This should almost always be used instead of []DishOccurrence.
	DishOccurrenceSlice []*DishOccurrence
	// DishOccurrenceHook is the signature for custom DishOccurrence hook methods
	DishOccurrenceHook func(context.Context, boil.ContextExecutor, *DishOccurrence) error

	dishOccurrenceQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	dishOccurrenceType                 = reflect.TypeOf(&DishOccurrence{})
	dishOccurrenceMapping              = queries.MakeStructMapping(dishOccurrenceType)
	dishOccurrencePrimaryKeyMapping, _ = queries.BindMapping(dishOccurrenceType, dishOccurrenceMapping, dishOccurrencePrimaryKeyColumns)
	dishOccurrenceInsertCacheMut       sync.RWMutex
	dishOccurrenceInsertCache          = make(map[string]insertCache)
	dishOccurrenceUpdateCacheMut       sync.RWMutex
	dishOccurrenceUpdateCache          = make(map[string]updateCache)
	dishOccurrenceUpsertCacheMut       sync.RWMutex
	dishOccurrenceUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var dishOccurrenceAfterSelectHooks []DishOccurrenceHook

var dishOccurrenceBeforeInsertHooks []DishOccurrenceHook
var dishOccurrenceAfterInsertHooks []DishOccurrenceHook

var dishOccurrenceBeforeUpdateHooks []DishOccurrenceHook
var dishOccurrenceAfterUpdateHooks []DishOccurrenceHook

var dishOccurrenceBeforeDeleteHooks []DishOccurrenceHook
var dishOccurrenceAfterDeleteHooks []DishOccurrenceHook

var dishOccurrenceBeforeUpsertHooks []DishOccurrenceHook
var dishOccurrenceAfterUpsertHooks []DishOccurrenceHook

// doAfterSelectHooks executes all "after Select" hooks.
func (o *DishOccurrence) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *DishOccurrence) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *DishOccurrence) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *DishOccurrence) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *DishOccurrence) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *DishOccurrence) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *DishOccurrence) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *DishOccurrence) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *DishOccurrence) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range dishOccurrenceAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddDishOccurrenceHook registers your hook function for all future operations.
func AddDishOccurrenceHook(hookPoint boil.HookPoint, dishOccurrenceHook DishOccurrenceHook) {
	switch hookPoint {
	case boil.AfterSelectHook:
		dishOccurrenceAfterSelectHooks = append(dishOccurrenceAfterSelectHooks, dishOccurrenceHook)
	case boil.BeforeInsertHook:
		dishOccurrenceBeforeInsertHooks = append(dishOccurrenceBeforeInsertHooks, dishOccurrenceHook)
	case boil.AfterInsertHook:
		dishOccurrenceAfterInsertHooks = append(dishOccurrenceAfterInsertHooks, dishOccurrenceHook)
	case boil.BeforeUpdateHook:
		dishOccurrenceBeforeUpdateHooks = append(dishOccurrenceBeforeUpdateHooks, dishOccurrenceHook)
	case boil.AfterUpdateHook:
		dishOccurrenceAfterUpdateHooks = append(dishOccurrenceAfterUpdateHooks, dishOccurrenceHook)
	case boil.BeforeDeleteHook:
		dishOccurrenceBeforeDeleteHooks = append(dishOccurrenceBeforeDeleteHooks, dishOccurrenceHook)
	case boil.AfterDeleteHook:
		dishOccurrenceAfterDeleteHooks = append(dishOccurrenceAfterDeleteHooks, dishOccurrenceHook)
	case boil.BeforeUpsertHook:
		dishOccurrenceBeforeUpsertHooks = append(dishOccurrenceBeforeUpsertHooks, dishOccurrenceHook)
	case boil.AfterUpsertHook:
		dishOccurrenceAfterUpsertHooks = append(dishOccurrenceAfterUpsertHooks, dishOccurrenceHook)
	}
}

// One returns a single dishOccurrence record from the query.
func (q dishOccurrenceQuery) One(ctx context.Context, exec boil.ContextExecutor) (*DishOccurrence, error) {
	o := &DishOccurrence{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboilerPSQL: failed to execute a one query for dish_occurrences")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all DishOccurrence records from the query.
func (q dishOccurrenceQuery) All(ctx context.Context, exec boil.ContextExecutor) (DishOccurrenceSlice, error) {
	var o []*DishOccurrence

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "sqlboilerPSQL: failed to assign all query results to DishOccurrence slice")
	}

	if len(dishOccurrenceAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all DishOccurrence records in the query.
func (q dishOccurrenceQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: failed to count dish_occurrences rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q dishOccurrenceQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "sqlboilerPSQL: failed to check if dish_occurrences exists")
	}

	return count > 0, nil
}

// Dish pointed to by the foreign key.
func (o *DishOccurrence) Dish(mods ...qm.QueryMod) dishQuery {
	queryMods := []qm.QueryMod{
		qm.Where("\"id\" = ?", o.DishID),
	}

	queryMods = append(queryMods, mods...)

	return Dishes(queryMods...)
}

// LoadDish allows an eager lookup of values, cached into the
// loaded structs of the objects. This is for an N-1 relationship.
func (dishOccurrenceL) LoadDish(ctx context.Context, e boil.ContextExecutor, singular bool, maybeDishOccurrence interface{}, mods queries.Applicator) error {
	var slice []*DishOccurrence
	var object *DishOccurrence

	if singular {
		var ok bool
		object, ok = maybeDishOccurrence.(*DishOccurrence)
		if !ok {
			object = new(DishOccurrence)
			ok = queries.SetFromEmbeddedStruct(&object, &maybeDishOccurrence)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", object, maybeDishOccurrence))
			}
		}
	} else {
		s, ok := maybeDishOccurrence.(*[]*DishOccurrence)
		if ok {
			slice = *s
		} else {
			ok = queries.SetFromEmbeddedStruct(&slice, maybeDishOccurrence)
			if !ok {
				return errors.New(fmt.Sprintf("failed to set %T from embedded struct %T", slice, maybeDishOccurrence))
			}
		}
	}

	args := make([]interface{}, 0, 1)
	if singular {
		if object.R == nil {
			object.R = &dishOccurrenceR{}
		}
		args = append(args, object.DishID)

	} else {
	Outer:
		for _, obj := range slice {
			if obj.R == nil {
				obj.R = &dishOccurrenceR{}
			}

			for _, a := range args {
				if a == obj.DishID {
					continue Outer
				}
			}

			args = append(args, obj.DishID)

		}
	}

	if len(args) == 0 {
		return nil
	}

	query := NewQuery(
		qm.From(`dishes`),
		qm.WhereIn(`dishes.id in ?`, args...),
	)
	if mods != nil {
		mods.Apply(query)
	}

	results, err := query.QueryContext(ctx, e)
	if err != nil {
		return errors.Wrap(err, "failed to eager load Dish")
	}

	var resultSlice []*Dish
	if err = queries.Bind(results, &resultSlice); err != nil {
		return errors.Wrap(err, "failed to bind eager loaded slice Dish")
	}

	if err = results.Close(); err != nil {
		return errors.Wrap(err, "failed to close results of eager load for dishes")
	}
	if err = results.Err(); err != nil {
		return errors.Wrap(err, "error occurred during iteration of eager loaded relations for dishes")
	}

	if len(dishOccurrenceAfterSelectHooks) != 0 {
		for _, obj := range resultSlice {
			if err := obj.doAfterSelectHooks(ctx, e); err != nil {
				return err
			}
		}
	}

	if len(resultSlice) == 0 {
		return nil
	}

	if singular {
		foreign := resultSlice[0]
		object.R.Dish = foreign
		if foreign.R == nil {
			foreign.R = &dishR{}
		}
		foreign.R.DishOccurrences = append(foreign.R.DishOccurrences, object)
		return nil
	}

	for _, local := range slice {
		for _, foreign := range resultSlice {
			if local.DishID == foreign.ID {
				local.R.Dish = foreign
				if foreign.R == nil {
					foreign.R = &dishR{}
				}
				foreign.R.DishOccurrences = append(foreign.R.DishOccurrences, local)
				break
			}
		}
	}

	return nil
}

// SetDish of the dishOccurrence to the related item.
// Sets o.R.Dish to related.
// Adds o to related.R.DishOccurrences.
func (o *DishOccurrence) SetDish(ctx context.Context, exec boil.ContextExecutor, insert bool, related *Dish) error {
	var err error
	if insert {
		if err = related.Insert(ctx, exec, boil.Infer()); err != nil {
			return errors.Wrap(err, "failed to insert into foreign table")
		}
	}

	updateQuery := fmt.Sprintf(
		"UPDATE \"dish_occurrences\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, []string{"dish_id"}),
		strmangle.WhereClause("\"", "\"", 2, dishOccurrencePrimaryKeyColumns),
	)
	values := []interface{}{related.ID, o.ID}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, updateQuery)
		fmt.Fprintln(writer, values)
	}
	if _, err = exec.ExecContext(ctx, updateQuery, values...); err != nil {
		return errors.Wrap(err, "failed to update local table")
	}

	o.DishID = related.ID
	if o.R == nil {
		o.R = &dishOccurrenceR{
			Dish: related,
		}
	} else {
		o.R.Dish = related
	}

	if related.R == nil {
		related.R = &dishR{
			DishOccurrences: DishOccurrenceSlice{o},
		}
	} else {
		related.R.DishOccurrences = append(related.R.DishOccurrences, o)
	}

	return nil
}

// DishOccurrences retrieves all the records using an executor.
func DishOccurrences(mods ...qm.QueryMod) dishOccurrenceQuery {
	mods = append(mods, qm.From("\"dish_occurrences\""))
	q := NewQuery(mods...)
	if len(queries.GetSelect(q)) == 0 {
		queries.SetSelect(q, []string{"\"dish_occurrences\".*"})
	}

	return dishOccurrenceQuery{q}
}

// FindDishOccurrence retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindDishOccurrence(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*DishOccurrence, error) {
	dishOccurrenceObj := &DishOccurrence{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from \"dish_occurrences\" where \"id\"=$1", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, dishOccurrenceObj)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "sqlboilerPSQL: unable to select from dish_occurrences")
	}

	if err = dishOccurrenceObj.doAfterSelectHooks(ctx, exec); err != nil {
		return dishOccurrenceObj, err
	}

	return dishOccurrenceObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *DishOccurrence) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboilerPSQL: no dish_occurrences provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(dishOccurrenceColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	dishOccurrenceInsertCacheMut.RLock()
	cache, cached := dishOccurrenceInsertCache[key]
	dishOccurrenceInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			dishOccurrenceAllColumns,
			dishOccurrenceColumnsWithDefault,
			dishOccurrenceColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(dishOccurrenceType, dishOccurrenceMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(dishOccurrenceType, dishOccurrenceMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO \"dish_occurrences\" (\"%s\") %%sVALUES (%s)%%s", strings.Join(wl, "\",\""), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO \"dish_occurrences\" %sDEFAULT VALUES%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			queryReturning = fmt.Sprintf(" RETURNING \"%s\"", strings.Join(returnColumns, "\",\""))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}

	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}

	if err != nil {
		return errors.Wrap(err, "sqlboilerPSQL: unable to insert into dish_occurrences")
	}

	if !cached {
		dishOccurrenceInsertCacheMut.Lock()
		dishOccurrenceInsertCache[key] = cache
		dishOccurrenceInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the DishOccurrence.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *DishOccurrence) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	dishOccurrenceUpdateCacheMut.RLock()
	cache, cached := dishOccurrenceUpdateCache[key]
	dishOccurrenceUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			dishOccurrenceAllColumns,
			dishOccurrencePrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("sqlboilerPSQL: unable to update dish_occurrences, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE \"dish_occurrences\" SET %s WHERE %s",
			strmangle.SetParamNames("\"", "\"", 1, wl),
			strmangle.WhereClause("\"", "\"", len(wl)+1, dishOccurrencePrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(dishOccurrenceType, dishOccurrenceMapping, append(wl, dishOccurrencePrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to update dish_occurrences row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: failed to get rows affected by update for dish_occurrences")
	}

	if !cached {
		dishOccurrenceUpdateCacheMut.Lock()
		dishOccurrenceUpdateCache[key] = cache
		dishOccurrenceUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q dishOccurrenceQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to update all for dish_occurrences")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to retrieve rows affected for dish_occurrences")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o DishOccurrenceSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("sqlboilerPSQL: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), dishOccurrencePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE \"dish_occurrences\" SET %s WHERE %s",
		strmangle.SetParamNames("\"", "\"", 1, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), len(colNames)+1, dishOccurrencePrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to update all in dishOccurrence slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to retrieve rows affected all in update all dishOccurrence")
	}
	return rowsAff, nil
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *DishOccurrence) Upsert(ctx context.Context, exec boil.ContextExecutor, updateOnConflict bool, conflictColumns []string, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("sqlboilerPSQL: no dish_occurrences provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(dishOccurrenceColumnsWithDefault, o)

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	if updateOnConflict {
		buf.WriteByte('t')
	} else {
		buf.WriteByte('f')
	}
	buf.WriteByte('.')
	for _, c := range conflictColumns {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	dishOccurrenceUpsertCacheMut.RLock()
	cache, cached := dishOccurrenceUpsertCache[key]
	dishOccurrenceUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			dishOccurrenceAllColumns,
			dishOccurrenceColumnsWithDefault,
			dishOccurrenceColumnsWithoutDefault,
			nzDefaults,
		)

		update := updateColumns.UpdateColumnSet(
			dishOccurrenceAllColumns,
			dishOccurrencePrimaryKeyColumns,
		)

		if updateOnConflict && len(update) == 0 {
			return errors.New("sqlboilerPSQL: unable to upsert dish_occurrences, could not build update column list")
		}

		conflict := conflictColumns
		if len(conflict) == 0 {
			conflict = make([]string, len(dishOccurrencePrimaryKeyColumns))
			copy(conflict, dishOccurrencePrimaryKeyColumns)
		}
		cache.query = buildUpsertQueryPostgres(dialect, "\"dish_occurrences\"", updateOnConflict, ret, update, conflict, insert)

		cache.valueMapping, err = queries.BindMapping(dishOccurrenceType, dishOccurrenceMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(dishOccurrenceType, dishOccurrenceMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	if len(cache.retMapping) != 0 {
		err = exec.QueryRowContext(ctx, cache.query, vals...).Scan(returns...)
		if errors.Is(err, sql.ErrNoRows) {
			err = nil // Postgres doesn't return anything when there's no update
		}
	} else {
		_, err = exec.ExecContext(ctx, cache.query, vals...)
	}
	if err != nil {
		return errors.Wrap(err, "sqlboilerPSQL: unable to upsert dish_occurrences")
	}

	if !cached {
		dishOccurrenceUpsertCacheMut.Lock()
		dishOccurrenceUpsertCache[key] = cache
		dishOccurrenceUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single DishOccurrence record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *DishOccurrence) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("sqlboilerPSQL: no DishOccurrence provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), dishOccurrencePrimaryKeyMapping)
	sql := "DELETE FROM \"dish_occurrences\" WHERE \"id\"=$1"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to delete from dish_occurrences")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: failed to get rows affected by delete for dish_occurrences")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q dishOccurrenceQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("sqlboilerPSQL: no dishOccurrenceQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to delete all from dish_occurrences")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: failed to get rows affected by deleteall for dish_occurrences")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o DishOccurrenceSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(dishOccurrenceBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), dishOccurrencePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM \"dish_occurrences\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, dishOccurrencePrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: unable to delete all from dishOccurrence slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "sqlboilerPSQL: failed to get rows affected by deleteall for dish_occurrences")
	}

	if len(dishOccurrenceAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *DishOccurrence) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindDishOccurrence(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *DishOccurrenceSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := DishOccurrenceSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), dishOccurrencePrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT \"dish_occurrences\".* FROM \"dish_occurrences\" WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 1, dishOccurrencePrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "sqlboilerPSQL: unable to reload all in DishOccurrenceSlice")
	}

	*o = slice

	return nil
}

// DishOccurrenceExists checks if the DishOccurrence row exists.
func DishOccurrenceExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from \"dish_occurrences\" where \"id\"=$1 limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "sqlboilerPSQL: unable to check if dish_occurrences exists")
	}

	return exists, nil
}
