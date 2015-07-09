package modules

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coopernurse/gorp"
	_ "github.com/go-sql-driver/mysql"
	"github.com/revel/revel"
)

var DB DBWrapper //Provides global access to this

type DBWrapper struct {
	DbMap    *gorp.DbMap
	Exported bool
}

func InitDB() {
	host, _ := revel.Config.String("db.host")
	user, _ := revel.Config.String("db.user")
	pass, _ := revel.Config.String("db.pass")
	database, _ := revel.Config.String("db.name")

	dbc, err := sql.Open("mysql", user+":"+pass+"@tcp("+host+")/"+database+"?autocommit=true&parseTime=true")

	if err != nil {
		revel.INFO.Println("Error occurred while attempting to connect to Teamwork database:", err)
		return
	}

	DB.DbMap = &gorp.DbMap{Db: dbc, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
}

// All queries throughout the app are wrapped up in a nice little package so we can take
// a look at what is actually happening during a single request.  In order to view these
// queries and request logs (not persisted at this time), simply append ?_profiling=true
// to any API call and view the output.  For example /v1/xxx.json?_profiling=true
// The plan is to extract it from the codebase, wrap this up in a module and distribute it freely.
func (db *DBWrapper) Select(rc *revel.Controller, i interface{}, SQL string, args ...interface{}) ([]interface{}, error) {
	before := time.Now().UnixNano()

	ProcessParams(&SQL, args)

	result, err := db.DbMap.Select(i, SQL, args...)

	CheckErr(err, "Selection query"+SQL[:Min(20, len(SQL))], "SQL", map[string]interface{}{
		"SQL":     SQL,
		"Params":  args,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			SQL:       SQL,
			Params:    args,
			TimeTaken: timeTaken,
			ShortSQL:  SQL[:Min(50, len(SQL))],
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (db *DBWrapper) SelectOne(rc *revel.Controller, i interface{}, SQL string, args ...interface{}) error {
	before := time.Now().UnixNano()

	ProcessParams(&SQL, args)

	err := db.DbMap.SelectOne(i, SQL, args...)

	CheckErr(err, "Error selecting one...", "SQL", map[string]interface{}{
		"SQL":     SQL,
		"Params":  args,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			SQL:       SQL,
			Params:    args,
			TimeTaken: timeTaken,
			ShortSQL:  SQL[:Min(50, len(SQL))],
		})

		req.QueryTime += timeTaken
	}

	return err
}

func (db *DBWrapper) SelectInt(rc *revel.Controller, SQL string, args ...interface{}) (int64, error) {
	before := time.Now().UnixNano()

	ProcessParams(&SQL, args)

	result, err := db.DbMap.SelectInt(SQL, args...)

	CheckErr(err, "Selecting int "+SQL[:Min(20, len(SQL))], "SQL", map[string]interface{}{
		"SQL":     SQL,
		"Params":  args,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			SQL:       SQL,
			Params:    args,
			TimeTaken: timeTaken,
			ShortSQL:  SQL[:Min(50, len(SQL))],
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (db *DBWrapper) SelectStr(rc *revel.Controller, SQL string, args ...interface{}) (string, error) {
	before := time.Now().UnixNano()

	result, err := db.DbMap.SelectStr(SQL, args...)

	CheckErr(err, "Selecting string "+SQL[:Min(20, len(SQL))], "SQL", map[string]interface{}{
		"SQL":     SQL,
		"Params":  args,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			SQL:       SQL,
			Params:    args,
			TimeTaken: timeTaken,
			ShortSQL:  SQL[:Min(50, len(SQL))],
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (db *DBWrapper) Insert(rc *revel.Controller, list ...interface{}) error {
	before := time.Now().UnixNano()

	err := db.DbMap.Insert(list...)

	CheckErr(err, "Insertion query failed", "SQL", map[string]interface{}{
		"Params":  list,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			TimeTaken: timeTaken,
			ShortSQL:  fmt.Sprintf("Inserting %%v", list),
		})

		req.QueryTime += timeTaken
	}

	return err
}

func (db *DBWrapper) Exec(rc *revel.Controller, SQL string, list ...interface{}) (sql.Result, error) {
	before := time.Now().UnixNano()

	result, err := db.DbMap.Exec(SQL, list...)

	CheckErr(err, "Exec "+SQL[:Min(20, len(SQL))], "SQL", map[string]interface{}{
		"SQL":     SQL,
		"Params":  list,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			SQL:       SQL,
			Params:    list,
			TimeTaken: timeTaken,
			ShortSQL:  SQL[:Min(50, len(SQL))],
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (dbWrap *DBWrapper) Get(rc *revel.Controller, i interface{}, keys ...interface{}) (interface{}, error) {
	before := time.Now().UnixNano()

	if dbWrap == nil {
		panic("dbWrap == nil")
	}
	if dbWrap.DbMap == nil {
		panic("dbWrap.dbMap == nil")
	}

	result, err := dbWrap.DbMap.Get(i, keys...)

	CheckErr(err, "Get query", "SQL", map[string]interface{}{
		"Params":  i,
		"Keys":    keys,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			TimeTaken: timeTaken,
			ShortSQL:  fmt.Sprintf("Getting %T with keys %#v", i, keys),
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (db *DBWrapper) Delete(rc *revel.Controller, i ...interface{}) (int64, error) {
	before := time.Now().UnixNano()

	result, err := db.DbMap.Delete(i...)

	CheckErr(err, "Delete query", "SQL", map[string]interface{}{
		"Params":  i,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			TimeTaken: timeTaken,
			ShortSQL:  fmt.Sprintf("Deleting keys %#v", i),
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (db *DBWrapper) Update(rc *revel.Controller, list ...interface{}) (int64, error) {
	before := time.Now().UnixNano()

	result, err := db.DbMap.Update(list...)

	CheckErr(err, "Update query", "SQL", map[string]interface{}{
		"Params":  list,
		"Request": rc.Request,
	})

	timeTaken := (time.Now().UnixNano() - before) / 1000 / 1000

	req := Requests[rc]

	if req != nil {
		req.Queries = append(req.Queries, QueryLog{
			TimeTaken: timeTaken,
			ShortSQL:  fmt.Sprintf("Updating %#v", list),
		})

		req.QueryTime += timeTaken
	}

	return result, err
}

func (db *DBWrapper) Begin() (*gorp.Transaction, error) {
	return db.DbMap.Begin()
}

func (db *DBWrapper) AddTableWithName(i interface{}, name string) *gorp.TableMap {
	return db.DbMap.AddTableWithName(i, name)
}

func ProcessParams(SQL *string, originalParams interface{}) {
	if params, ok := originalParams.([]interface{}); ok {
		if len(params) == 1 {
			if paramMap, k := params[0].(map[string]interface{}); k {
				// Loop through each entry in the param map
				for key, value := range paramMap {
					if strSlice, isStrSlice := value.([]string); isStrSlice {
						var replacements []string

						for id, str := range strSlice {
							paramMap[key+strconv.Itoa(id)] = str

							replacements = append(replacements, ":"+key+strconv.Itoa(id))
						}

						*SQL = strings.Replace(*SQL, ":"+key, strings.Join(replacements, ","), -1)

						originalParams = &paramMap
					} else if intSlice, isIntSlice := value.([]int64); isIntSlice {
						var replacements []string

						for id, val64 := range intSlice {
							paramMap[key+strconv.Itoa(id)] = val64

							replacements = append(replacements, ":"+key+strconv.Itoa(id))
						}

						*SQL = strings.Replace(*SQL, ":"+key, strings.Join(replacements, ","), -1)

						originalParams = &paramMap
					}
				}
			}
		}
	}
}
