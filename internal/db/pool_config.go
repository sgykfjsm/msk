package db

import "time"

type PoolConfig struct {
	// MaxOpenConns is the maximum number of open connections to the database.
	MaxOpenConns int
	// MaxIdleConns is the maximum number of idle connections in the pool.
	MaxIdleConns int
	// ConnMaxLifetime is the maximum amount of time a connection may be reused.
	ConnMaxLifetime time.Duration
	// ConnMaxIdleTime is the maximum amount of time a connection may be idle.
	ConnMaxIdleTime time.Duration
}

// NewPoolConfig creates a new PoolConfig with default values.
func NewPoolConfig() *PoolConfig {
	pc := &PoolConfig{}
	pc.SetDefaults()
	return pc
}

func (c *PoolConfig) SetDefaults() {
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 10 // Default value for maximum open connections
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 5 // Default value for maximum idle connections
	}
	if c.ConnMaxLifetime <= 0 {
		c.ConnMaxLifetime = 60 * time.Second // Default value for connection max lifetime in seconds
	}
	if c.ConnMaxIdleTime <= 0 {
		c.ConnMaxIdleTime = 30 * time.Second // Default value for connection max idle time in seconds
	}
}
