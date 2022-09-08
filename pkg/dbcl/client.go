package dbcl

import (
	"fmt"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	maxIdleConns = 35
	maxOpenConns = 50

	defaultQs = "?multiStatements=true&parseTime=true&loc=UTC"
)

type Client struct {
	readClient  *sqlx.DB
	writeClient *sqlx.DB
	migrations  map[string]string
}

func (c *Client) Reader() *sqlx.DB {
	return c.readClient
}

func (c *Client) Writer() *sqlx.DB {
	return c.writeClient
}

func (c *Client) Ping() error {
	err := c.readClient.Ping()
	if err != nil {
		return fmt.Errorf("failed on read ping: %w", err)
	}

	err = c.writeClient.Ping()
	if err != nil {
		return fmt.Errorf("failed on write ping: %w", err)
	}

	return nil
}

func (c *Client) Begin() (*Tx, error) {
	tx, err := c.writeClient.Beginx()
	if err != nil {
		return nil, err
	}

	return NewTx(tx), nil
}

func (c *Client) Close() {
	c.readClient.Close()
	c.writeClient.Close()
}

func initConnection(dsn string) (*sqlx.DB, error) {
	client, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}

	client.DB.SetMaxIdleConns(maxIdleConns)
	client.DB.SetMaxOpenConns(maxOpenConns)

	return client, nil
}

func New(writeHost, readHost, port, name, user, pass string, migrations map[string]string) (*Client, error) {
	readDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", user, pass, readHost, port, name, defaultQs)
	readClient, err := initConnection(readDSN)
	if err != nil {
		return nil, err
	}

	writeDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s%s", user, pass, writeHost, port, name, defaultQs)
	writeClient, err := initConnection(writeDSN)
	if err != nil {
		return nil, err
	}

	client := &Client{
		readClient:  readClient,
		writeClient: writeClient,
		migrations:  migrations,
	}

	if err := client.Ping(); err != nil {
		client.Close()
		return nil, err
	}

	return client, nil
}
