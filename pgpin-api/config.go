package main

import (
	"github.com/darkhelmet/env"
	"github.com/fernet/fernet-go"
	"time"
)

var (
	ConfigDatabaseConnectTimeout   = 5 * time.Second
	ConfigDatabaseStatementTimeout = 5 * time.Second
	ConfigDatabasePoolSize         = 5
	ConfigDatabaseUrl              = env.String("DATABASE_URL")
	ConfigFernetKeys               = fernet.MustDecodeKeys(env.String("FERNET_KEYS"))
	ConfigFernetTtl                = time.Hour * 24 * 365 * 10
	ConfigPinRefreshInterval       = 20 * time.Minute
	ConfigPinResultsRowsMax        = 10000
	ConfigPinStatementTimeout      = 30 * time.Second
	ConfigRedisPoolSize            = 5
	ConfigRedisUrl                 = env.String("REDIS_URL")
	ConfigSchedulerTickInterval    = 10 * time.Second
	ConfigTestLogs                 = env.StringDefault("TEST_LOGS", "false") != "true"
	ConfigWebPort                  = env.IntDefault("PORT", 5000)
	ConfigWebTimeout               = time.Second * 10
)
