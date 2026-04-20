// Package database 数据库服务
// 提供 SQLite 数据库连接和操作功能
package database

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strconv"
	"time"

	"go-downloader/internal/config"
	"go-downloader/internal/utils/logger"

	_ "github.com/mattn/go-sqlite3"
)

// Service 数据库服务接口，定义所有数据库操作方法
type Service interface {
	Health() map[string]string                    // 健康检查，返回数据库连接状态
	Close() error                                 // 关闭数据库连接
	CreateItem(name string) (int64, error)        // 创建项目，返回新项目 ID
	GetItems() ([]map[string]interface{}, error)  // 查询所有项目
	GetItem(id int64) (map[string]interface{}, error) // 根据 ID 查询单个项目
	UpdateItem(id int64, name string) error       // 更新项目名称
	DeleteItem(id int64) error                    // 删除项目
}

// service 数据库服务实现
type service struct {
	db *sql.DB
}

// dbInstance 数据库服务单例
var dbInstance *service

// New 创建数据库服务单例，自动执行迁移
// 如果单例已存在则直接返回，否则创建新的数据库连接
func New() Service {
	if dbInstance != nil {
		return dbInstance
	}

	// 构建数据库文件路径
	dburl := filepath.Join(config.GetDataDir(), "go-downloader.db")

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		logger.Fatal("打开数据库失败: %v", err)
	}

	// 创建服务实例
	dbInstance = &service{
		db: db,
	}

	// 执行数据库迁移
	if err := dbInstance.migrate(); err != nil {
		logger.Fatal("数据库迁移失败: %v", err)
	}
	
	logger.Info("数据库已连接: %s", dburl)
	return dbInstance
}

// migrate 执行数据库迁移：创建表
// 创建 items 表用于存储项目数据
func (s *service) migrate() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// CreateItem 创建新项目，返回插入后的自增 ID
func (s *service) CreateItem(name string) (int64, error) {
	result, err := s.db.Exec("INSERT INTO items (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// GetItems 查询所有项目，按 ID 降序排列
func (s *service) GetItems() ([]map[string]interface{}, error) {
	rows, err := s.db.Query("SELECT id, name, created_at, updated_at FROM items ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, createdAt, updatedAt string
		if err := rows.Scan(&id, &name, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"id":         id,
			"name":       name,
			"created_at": createdAt,
			"updated_at": updatedAt,
		})
	}
	return items, nil
}

// GetItem 根据 ID 查询单个项目
func (s *service) GetItem(id int64) (map[string]interface{}, error) {
	var name, createdAt, updatedAt string
	err := s.db.QueryRow("SELECT name, created_at, updated_at FROM items WHERE id = ?", id).Scan(&name, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"id":         id,
		"name":       name,
		"created_at": createdAt,
		"updated_at": updatedAt,
	}, nil
}

// UpdateItem 更新指定 ID 的项目名称，并刷新更新时间
func (s *service) UpdateItem(id int64, name string) error {
	_, err := s.db.Exec("UPDATE items SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", name, id)
	return err
}

// DeleteItem 根据 ID 删除项目
func (s *service) DeleteItem(id int64) error {
	_, err := s.db.Exec("DELETE FROM items WHERE id = ?", id)
	return err
}

// Health 检查数据库连接健康状态，返回连接池统计信息
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// 执行 Ping 检查连接
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		logger.Error("数据库连接异常: %v", err)
		return stats
	}

	// 连接正常
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// 获取连接池统计信息
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// 检查连接池健康状态并给出警告
	if dbStats.OpenConnections > 40 {
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close 关闭数据库连接
func (s *service) Close() error {
	logger.Info("数据库已断开: %s", filepath.Join(config.GetDataDir(), "go-downloader.db"))
	return s.db.Close()
}
