package repositories

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"mt/config"
	"mt/pkg/cache"
	"mt/pkg/db"
	"mt/pkg/logger"
)

var _ DataRepo = (*dataRepo)(nil)

type DataRepo interface {
	DB(name string) db.Db
	DbRepo() DbRepo
	Redis(name string) cache.Redis
	RedisRepo() RedisRepo
}

type dataRepo struct {
	db    *dbRepo
	redis *redisRepo
}

func NewDataRepo(logger *logger.Logger, conf *config.Data) DataRepo {
	var (
		ctx = context.Background()

		dbRepo    = new(dbRepo)
		redisRepo = new(redisRepo)
		repo      = new(dataRepo)

		dbMap    = make(map[string]*config.DatabaseItem, 1)
		redisMap = make(map[string]*config.RedisItem, 1)
	)

	// 初始化数据库
	var database = conf.GetDatabase()
	if database.GetDefault() != nil {
		dbMap[DbConnectionDefaultName] = database.GetDefault()
	}

	lenDatabase := len(dbMap)
	if lenDatabase > 0 {
		dbRepo.resource = make(map[string]db.Db, lenDatabase)
		for dbName, dbConfig := range dbMap {
			rdb, err := db.NewDb(dbName, dbConfig, logger)
			if err != nil {
				logger.UseApp(ctx).Error(fmt.Sprintf("init db.repo %s error", dbName), zap.Error(err))
			} else {
				logger.UseApp(ctx).Info(fmt.Sprintf("init db.repo %s successfully", dbName))
				dbRepo.resource[dbName] = rdb
			}
		}

		repo.db = dbRepo
	} else {
		logger.UseApp(ctx).Warn("Currently not db.repo connected.")
	}

	// 初始化 Redis
	var redis = conf.GetRedis()
	if redis.GetDefault() != nil {
		redisMap[RedisConnectionDefaultName] = redis.GetDefault()
	}

	lenRedis := len(redisMap)
	if lenRedis > 0 {
		redisRepo.resource = make(map[string]cache.Redis, lenRedis)
		for redisName, redisConfig := range redisMap {
			redis, err := cache.NewRedis(redisName, redisConfig)
			if err != nil {
				logger.UseApp(ctx).Error(fmt.Sprintf("init redis.repo %s error", redisName), zap.Error(err))
			} else {
				logger.UseApp(ctx).Info(fmt.Sprintf("init redis.repo %s successfully", redisName))
				redisRepo.resource[redisName] = redis
			}
		}

		repo.redis = redisRepo
	} else {
		logger.UseApp(ctx).Warn("Currently not redis.repo connected.")
	}

	return repo
}

func (repo *dataRepo) DB(name string) db.Db {
	return repo.db.resource[name]
}

func (repo *dataRepo) DbRepo() DbRepo {
	return repo.db
}

func (repo *dataRepo) Redis(name string) cache.Redis {
	return repo.redis.resource[name]
}

func (repo *dataRepo) RedisRepo() RedisRepo {
	return repo.redis
}
