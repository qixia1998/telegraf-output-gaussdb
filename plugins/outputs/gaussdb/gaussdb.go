//go:generate ../../../tools/readme_config_includer/generator
package gaussdb

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	_ "gitee.com/opengauss/openGauss-connector-go-pq"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"log"
)

//go:embed sample.conf
var sampleConfig string

// GaussDB is the top level struct for this plugin.
type GaussDB struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Dbname   string `toml:"dbname"`
	Table    string `toml:"table"`
	conn     *sql.DB
	Debug    bool `toml:"debug"`
}

// SampleConfig returns the sample config for this plugin
func (*GaussDB) SampleConfig() string {
	return sampleConfig
}

// Init is for setup, and validating config
func (p *GaussDB) Init() error {
	if p.Host == "" {
		return fmt.Errorf("host is not null")
	}

	if p.User == "" || p.Password == "" {
		return fmt.Errorf("user or password failed")
	}

	if p.Dbname == "" {
		p.Dbname = "postgres"
	}

	if p.Table == "" {
		return fmt.Errorf("table is not null")
	}

	if p.Debug {
		log.Println("Successfully initialized GaussDB output plugin")
	}

	return nil
}

// Connect GaussDB
func (p *GaussDB) Connect() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Dbname)

	var err error
	p.conn, err = sql.Open("opengauss", connStr)
	if err != nil {
		return fmt.Errorf("unable to connect to GaussDB: %v", err)
	}

	// Check GaussDB
	err = p.conn.Ping()
	if err != nil {
		return fmt.Errorf("unable to ping GaussDB: %v", err)
	}

	// Create Table
	createTableQuery := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS %s (
      timestamp TIMESTAMPTZ NOT NULL,
      measurement TEXT NOT NULL,
      tags JSONB NOT NULL,
      fields JSONB NOT NULL
    )`, p.Table)
	_, err = p.conn.Exec(createTableQuery)
	if err != nil {
		return fmt.Errorf("unable to Create Table: %v", err)
	}
	if p.Debug {
		log.Println("Successfully connected to GaussDB")
	}
	return nil
}

// Write writes metrics to GaussDB
func (p *GaussDB) Write(metrics []telegraf.Metric) error {
	tx, err := p.conn.Begin()
	if err != nil {
		return fmt.Errorf("unable to start transaction: %v", err)
	}
	defer tx.Rollback()

	insertQuery := fmt.Sprintf(`
    INSERT INTO %s (timestamp, measurement, tags, fields) 
    VALUES ($1, $2, $3, $4)`, p.Table)

	for _, metric := range metrics {
		tags, err := json.Marshal(metric.Tags())
		if err != nil {
			return fmt.Errorf("unable to serialise tags: %v", err)
		}
		fields, err := json.Marshal(metric.Fields())
		if err != nil {
			return fmt.Errorf("unable to serialise field: %v", err)
		}

		_, err = tx.Exec(insertQuery, metric.Time().UTC(), metric.Name(), tags, fields)
		if err != nil {
			return fmt.Errorf("failed to insert data: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("unable to submit transaction: %v", err)
	}

	log.Println("Data written successfully")
	return nil
}

// Close is a no-op for this plugin
func (p *GaussDB) Close() error {
	return nil
}

func init() {
	outputs.Add("gauss", func() telegraf.Output { return &GaussDB{} })
}
