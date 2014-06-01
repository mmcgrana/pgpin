package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bmizerany/pq"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type pinSlim struct {
	Id         string `json:"id"`
	ResourceId string `json:"resource_id"`
	Name       string `json:"name"`
}

type pin struct {
	Id                string     `json:"id"`
	ResourceId        string     `json:"resource_id"`
	Name              string     `json:"name"`
	Sql               string     `json:"sql"`
	UserId            string     `json:"user_id"`
	CreatedAt         time.Time  `json:"created_at"`
	ResourceUrl       *string    `json:"-"`
	ResultsFieldsJson *string    `json:"results_fields_json"`
	ResultsRowsJson   *string    `json:"results_rows_json"`
	ErrorMessage      *string    `json:"error_message"`
	QueryStartedAt    *time.Time `json:"query_started_at"`
	QueryFinishedAt   *time.Time `json:"query_finished_at"`
	DeletedAt         *time.Time `json:"-"`
	LockSeq           int        `json:"-"`
}

type attachment struct {
	AppName   string `json:"app_name"`
	ConfigVar string `json:"config_var"`
}

type resource struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Url         string       `json:"-"`
	Attachments []attachment `json:"attachments"`
}

type notFoundError struct {
	Message string
}
func (e notFoundError) Error() string {
    return e.Message
}

type notAuthorizedError struct {
	Message string
}
func (e notAuthorizedError) Error() string {
	return e.Message
}

type malformedError struct {
	Message string
}
func (e malformedError) Error() string {
	return e.Message
}

type invalidError struct {
	Message string
}
func (e invalidError) Error() string {
	return e.Message
}

func dataMustParseDatabaseUrl(s string) string {
	conf, err := pq.ParseURL(s)
	if err != nil {
		panic(err)
	}
	return conf
}

var db *sql.DB

func dataInit() {
	log("key=data.init.start")
	dbConf := dataMustParseDatabaseUrl(mustGetenv("DATABASE_URL"))
	dbNew, err := sql.Open("postgres", dbConf)
	if err != nil {
		panic(err)
	}
	db = dbNew
	log("key=data.init.finish")
}

func dataTest() error {
	var r int
	err := db.QueryRow("SELECT 1").Scan(&r)
	if err != nil {
		return err
	}
	return nil
}

func dataGetUserId(token string) (string, error) {
	log("key=data.get_user_id.start")
	resp, err := http.Get("https://:" + token + "@api.heroku.com/account")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return "", notAuthorizedError{Message: "not authorized"}
	} else if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf("status code %d", resp.StatusCode))
	}
	var val map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&val)
	if err != nil {
		return "", err
	}
	log("key=data.get_user_id.finish")
	return val["id"].(string), nil
}

func dataGetResources(token string) ([]resource, error) {
	log("key=data.get_resources.start")
	resp, err := http.Get("https://:" + token + "@api.heroku.com/resources")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 401 {
		return nil, notAuthorizedError{Message: "not authorized"}
	} else if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("status code %d", resp.StatusCode))
	}
	var vals []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&vals)
	if err != nil {
		return nil, err
	}
	resources := []resource{}
	for _, val := range vals {
		id := val["id"].(string)
		name := val["name"].(string)
		url := val["value"].(string)
		attachments := []attachment{}
		for _, a := range val["attachments"].([]interface{}) {
			at := a.(map[string]interface{})
			appName := at["app"].(map[string]interface{})["name"].(string)
			configVar := at["config_var"].(string)
			attachment := attachment{appName, configVar}
			attachments = append(attachments, attachment)
		}
		resource := resource{id, name, url, attachments}
		resources = append(resources, resource)
	}
	log("key=data.get_resources.finish")
	return resources, nil
}

func dataGetResource(token string, resourceId string) (*resource, error) {
	resources, err := dataGetResources(token)
	if err != nil {
		return nil, err
	}
	for _, resource := range resources {
		if resource.Id == resourceId {
			return &resource, nil
		}
	}
	return nil, notFoundError{Message: "resource not found"}
}

func dataGetUserIdAndResource(token string, resourceId string) (string, *resource, error) {
	userId, err := dataGetUserId(token)
	if err != nil {
		return "", nil, err
	}
	resources, err := dataGetResources(token)
	if err != nil {
		return "", nil, err
	}
	for _, resource := range resources {
		if resource.Id == resourceId {
			return userId, &resource, nil
		}
	}
	return "", nil, notFoundError{Message: "resource not found"}
}

func dataGetPins(token string) ([]pinSlim, error) {
	resources, err := dataGetResources(token)
	if err != nil {
		return nil, err
	}
	resourceIds := []string{}
	for _, resource := range resources {
		resourceIds = append(resourceIds, resource.Id)
	}
	res, err := db.Query("SELECT id, resource_id, name FROM pins WHERE deleted_at IS NULL and resource_id in ('" + strings.Join(resourceIds, "','") + "')")
	if err != nil {
		return nil, err
	}
	defer res.Close()
	pins := []pinSlim{}
	for res.Next() {
		pin := pinSlim{}
		err = res.Scan(&pin.Id, &pin.ResourceId, &pin.Name)
		if err != nil {
			return nil, err
		}
		pins = append(pins, pin)
	}
	return pins, nil
}

var emptyRegexp = regexp.MustCompile("\\A\\s*\\z")

func dataValidateNonempty(f string, s string) error {
	if emptyRegexp.MatchString(s) {
		return invalidError{Message: fmt.Sprintf("field %s must be nonempty", f)}
	}
	return nil
}

func dataCreatePin(token, resourceId, name, sql string) (*pin, error) {
	if err := dataValidateNonempty("name", name); err != nil {
		return nil, err
	}
	if err := dataValidateNonempty("sql", sql); err != nil {
		return nil, err
	}
	userId, resource, err := dataGetUserIdAndResource(token, resourceId)
	if resource == nil {
		return nil, notFoundError{Message: "resource not found"}
	}
	pin := pin{}
	pin.Id = randUuid()
	pin.ResourceId = resourceId
	pin.Name = name
	pin.Sql = sql
	pin.UserId = userId
	pin.CreatedAt = time.Now()
	pin.ResourceUrl = &resource.Url
	pin.LockSeq = 1
	_, err = db.Exec("INSERT into pins (id, resource_id, name, sql, user_id, created_at, resource_url, lock_seq) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		pin.Id, pin.ResourceId, pin.Name, pin.Sql, pin.UserId, pin.CreatedAt, pin.ResourceUrl, pin.LockSeq)
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

func dataGetPinInternal(queryFrag string, queryVals ...interface{}) (*pin, error) {
	res, err := db.Query(`SELECT id, resource_id, name, sql, user_id, created_at, resource_url, results_fields_json, results_rows_json, error_message, query_started_at, query_finished_at, deleted_at, lock_seq
    	                  FROM pins `+queryFrag+` LIMIT 1`, queryVals...)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	ok := res.Next()
	if !ok {
		return nil, nil
	}
	pin := pin{}
	err = res.Scan(&pin.Id, &pin.ResourceId, &pin.Name, &pin.Sql, &pin.UserId, &pin.CreatedAt, &pin.ResourceUrl, &pin.ResultsFieldsJson, &pin.ResultsRowsJson, &pin.ErrorMessage, &pin.QueryStartedAt, &pin.QueryFinishedAt, &pin.DeletedAt, &pin.LockSeq)
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

func dataGetPin(token string, id string) (*pin, error) {
	pin, err := dataGetPinInternal("WHERE id=$1 AND deleted_at IS NULL", id)
	if err != nil {
		return nil, err
	}
	if pin == nil {
		return nil, notFoundError{Message: "pin not found"}
	}
	resource, err := dataGetResource(token, pin.ResourceId)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, notFoundError{Message: "resource not found"}
	}
	return pin, nil
}

func dataUpdatePin(pin *pin) (error) {
	res, err := db.Exec("UPDATE pins SET resource_id=$1, name=$2, sql=$3, user_id=$4, created_at=$5, resource_url=$6, results_fields_json=$7, results_rows_json=$8, error_message=$9, query_started_at=$10, query_finished_at=$11, deleted_at=$12, lock_seq=$13 WHERE id=$14 and lock_seq=$15",
		pin.ResourceId, pin.Name, pin.Sql, pin.UserId, pin.CreatedAt, pin.ResourceUrl, pin.ResultsFieldsJson, pin.ResultsRowsJson, pin.ErrorMessage, pin.QueryStartedAt, pin.QueryFinishedAt, pin.DeletedAt, pin.LockSeq+1, pin.Id, pin.LockSeq)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("optimistic lock error")
	}
	pin.LockSeq += 1
	return nil
}

func dataDeletePin(pin *pin) (error) {
	deletedAt := time.Now()
	pin.DeletedAt = &deletedAt
	return dataUpdatePin(pin)
}
