package client

import (
	"fmt"
)

type DatabaseEngine string

const (
	EngineAmazonRedshift   DatabaseEngine = "redshift"
	EngineBigQuery         DatabaseEngine = "bigquery"
	EngineDruid            DatabaseEngine = "druid"
	EngineGoogleAnalytics  DatabaseEngine = "googleanalytics"
	EngineH2               DatabaseEngine = "h2"
	EngineMongoDB          DatabaseEngine = "mongo"
	EngineMySQL            DatabaseEngine = "mysql"
	EngineOracle           DatabaseEngine = "oracle"
	EnginePostgres         DatabaseEngine = "postgres"
	EnginePresto           DatabaseEngine = "presto-jdbc"
	EnginePrestoDeprecated DatabaseEngine = "presto"
	EngineSnowflake        DatabaseEngine = "snowflake"
	EngineSparkSQL         DatabaseEngine = "sparksql"
	EngineSQLServer        DatabaseEngine = "sqlserver"
	EngineSQLite           DatabaseEngine = "sqlite"
)

var DatabaseAllowedEngines = []DatabaseEngine{
	EngineAmazonRedshift,
	EngineBigQuery,
	EngineDruid,
	EngineGoogleAnalytics,
	EngineH2,
	EngineMongoDB,
	EngineMySQL,
	EngineOracle,
	EnginePostgres,
	EnginePresto,
	EnginePrestoDeprecated,
	EngineSnowflake,
	EngineSparkSQL,
	EngineSQLServer,
	EngineSQLite,
}

type DatabaseDetails map[string]interface{}
type DatabaseSchedule struct {
	Day    *string `json:"schedule_day"`
	Frame  *string `json:"schedule_frame"`
	Hour   *int64  `json:"schedule_hour"`
	Minute *int64  `json:"schedule_minute"`
	Type   string  `json:"schedule_type"`
}

type Database struct {
	Id          int64   `json:"id"`
	Name        string  `json:"name"`
	Engine      string  `json:"engine"`
	Description *string `json:"description"`
	Caveats     *string `json:"caveats"`
	Timezone    string  `json:"timezone"`

	CreatedAt string  `json:"created_at"`
	CreatorID *int64  `json:"creator_id"`
	UpdatedAt *string `json:"updated_at"`

	AutoRunQueries bool `json:"auto_run_queries"`
	CanManage      bool `json:"can_manage"`
	IsFullSync     bool `json:"is_full_sync"`
	IsSample       bool `json:"is_sample"`
	IsOnDemand     bool `json:"is_on_demand"`

	Features  []string                     `json:"features"`
	Details   *DatabaseDetails             `json:"details"`
	Options   *map[string]string           `json:"options"`
	Settings  *map[string]string           `json:"settings"`
	Schedules *map[string]DatabaseSchedule `json:"schedules"`

	CacheFieldValuesSchedule *string `json:"cache_field_values_schedule"`
	CacheTTL                 *string `json:"cache_ttl"`
}

type DatabaseRequest struct {
	Engine    DatabaseEngine               `json:"engine"`
	Name      string                       `json:"name"`
	Details   DatabaseDetails              `json:"details"`
	Schedules *map[string]DatabaseSchedule `json:"schedules"`
}

func (c *Client) GetDatabase(databaseId int64) (*Database, error) {
	var database Database
	err := c.doGet(fmt.Sprintf("/database/%d", databaseId), &database)
	if err != nil {
		return nil, err
	}

	return &database, nil
}

func (c *Client) CreateDatabase(request DatabaseRequest) (int64, error) {
	var database Database
	err := c.doPost("/database", request, &database)
	if err != nil {
		return 0, err
	}

	return database.Id, nil
}

func (c *Client) UpdateDatabase(databaseId int64, request DatabaseRequest) error {
	var database Database
	err := c.doPut(fmt.Sprintf("/database/%d", databaseId), request, &database)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteDatabase(databaseId int64) error {
	return c.doDelete(fmt.Sprintf("/database/%d", databaseId), nil)
}
