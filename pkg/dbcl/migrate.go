package dbcl

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func (c *Client) initMigrationTable() error {
	_, err := c.writeClient.Exec(`CREATE TABLE IF NOT EXISTS 
				migrations(id VARCHAR(100) PRIMARY KEY)
				AS SELECT "" AS id;`)
	return err
}

func (c *Client) migrationStatus() (string, error) {
	var id string

	err := c.initMigrationTable()
	if err != nil {
		return id, err
	}

	err = c.readClient.QueryRow("SELECT id FROM migrations").Scan(&id)
	if err != nil {
		return id, err
	}

	return id, nil
}

func (c *Client) updateMigrationStatus(status string) error {
	_, err := c.writeClient.Exec("UPDATE migrations SET id = ?", status)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) getMigrationList() ([]string, error) {
	var files []string
	for k := range c.migrations {
		if !strings.HasSuffix(k, ".down.sql") {
			files = append(files, k)
		}
	}

	return files, nil
}

func parseVersion(path string) (int, error) {
	s := strings.Split(path, "_")
	version, err := strconv.Atoi(s[0])
	if err != nil {
		return version, err
	}

	return version, nil
}

func (c *Client) getNextMigrations() ([]string, error) {
	var versions []string

	current, err := c.migrationStatus()
	if err != nil {
		return versions, err
	}

	files, err := c.getMigrationList()
	if err != nil {
		return versions, err
	}

	currentVersion := -1
	if len(current) > 0 {
		currentVersion, err = parseVersion(current)
		if err != nil {
			return versions, err
		}
	}

	for _, path := range files {
		v, err := parseVersion(path)
		if err != nil {
			return versions, err
		}

		if v > currentVersion {
			versions = append(versions, path)
		}
	}

	return versions, nil
}

func (c *Client) getPrevMigration() (string, error) {
	var version string

	current, err := c.migrationStatus()
	if err != nil {
		return version, err
	}

	files, err := c.getMigrationList()
	if err != nil {
		return version, err
	}

	currentVersion, err := parseVersion(current)
	if err != nil {
		return version, err
	}

	highestVersion := -1

	for _, path := range files {
		v, err := parseVersion(path)
		if err != nil {
			return version, err
		}

		if v < currentVersion && v > highestVersion {
			highestVersion = v
			version = path
		}
	}

	return version, nil
}

func (c *Client) UpgradeMigrations() error {
	migrations, err := c.getNextMigrations()
	if err != nil {
		return err
	}

	if len(migrations) > 0 {
		sort.Strings(migrations)

		for _, migration := range migrations {
			err = c.execMigration(migration)
			if err != nil {
				return fmt.Errorf("%s: %w", migration, err)
			}
		}

		c.updateMigrationStatus(migrations[len(migrations)-1])
	}

	return nil
}

func (c *Client) DowngradeMigration() error {
	current, err := c.migrationStatus()
	if err != nil {
		return err
	}

	if len(current) > 0 {
		down := strings.ReplaceAll(current, ".sql", ".down.sql")
		err := c.execMigration(down)
		if err != nil {
			return fmt.Errorf("%s: %w", down, err)
		}

		prev, err := c.getPrevMigration()
		if err != nil {
			return err
		}

		c.updateMigrationStatus(prev)
	}

	return nil
}

func (c *Client) DowngradeMigrations() error {
	for {
		current, err := c.migrationStatus()
		if err != nil {
			return err
		}

		if len(current) == 0 {
			break
		}

		err = c.DowngradeMigration()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) execMigration(path string) error {
	text, ok := c.migrations[path]
	if !ok {
		return fmt.Errorf("invalid migration")
	}

	_, err := c.writeClient.Exec(string(text))
	if err != nil {
		return err
	}

	return nil
}
