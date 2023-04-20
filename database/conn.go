package database

//import (
//	"strconv"
//	"strings"
//	"time"
//
//	"github.com/go-sql-driver/mysql"
//	mysqldb "gorm.io/driver/mysql"
//	"gorm.io/gorm"
//)
//
///*
//kubectl patch svc -n resource mariadb-mariadb-cluster -p '{"spec": {"type": "NodePort"}}'
//kubectl get svc -n resource -o go-template='{{range .items}}{{if eq .metadata.name "mariadb-mariadb-cluster"}}{{range.spec.ports}}{{if .nodePort}}{{.nodePort}}{{"\n"}}{{end}}{{end}}{{end}}{{end}}'
//kubectl get svc -n resource mariadb-mariadb-cluster -o go-template='{{range.spec.ports}}{{if .nodePort}}{{.nodePort}}{{end}}{{end}}'
//kubectl get svc -n resource mariadb-mariadb-cluster -o go-template='{{range.spec.ports}}{{if .nodePort}}{{.nodePort}}{{"\n"}}{{end}}{{end}}'
//*/
//
//var db *gorm.DB
//
//// Config 数据库配置 TODO rename
//type Config struct {
//	Host            string
//	Port            string
//	DataBaseName    string
//	User            string
//	Password        string
//	Timezone        string
//	MaxIdleConns    string
//	MaxOpenConns    string
//	ConnMaxLifetime string
//}
//
//// InitGormDB return *gorm.DB
//func InitGormDB(config *Config) error {
//	var gormDB *gorm.DB
//	var err error
//	location, err := time.LoadLocation(config.Timezone)
//	if err != nil {
//		return err
//	}
//	mysqlConfig := mysql.NewConfig()
//	if find := strings.Contains(config.Host, ":"); find {
//		mysqlConfig.Addr = "[" + config.Host + "]" + ":" + config.Port
//	} else {
//		mysqlConfig.Addr = config.Host + ":" + config.Port
//	}
//	mysqlConfig.User = config.User
//	mysqlConfig.Passwd = config.Password
//	mysqlConfig.DBName = config.DataBaseName
//	mysqlConfig.Net = "tcp"
//	mysqlConfig.Loc = location
//	mysqlConfig.Params = map[string]string{
//		"charset":   "utf8mb4",
//		"parseTime": "true",
//	}
//	dsn := mysqlConfig.FormatDSN()
//	if gormDB, err = gorm.Open(mysqldb.Open(dsn)); err != nil {
//		return err
//	}
//
//	maxIdleConns, err := strconv.Atoi(config.MaxIdleConns)
//	if err != nil {
//		return err
//	}
//
//	maxOpenConns, err := strconv.Atoi(config.MaxOpenConns)
//	if err != nil {
//		return err
//	}
//
//	connMaxLifetime, err := time.ParseDuration(config.ConnMaxLifetime)
//	if err != nil {
//		return err
//	}
//
//	if sqlDB, err := gormDB.DB(); err != nil {
//		return err
//	} else {
//		/*
//			SetMaxIdleConns设置空闲连接池的最大连接数。
//
//			如果MaxOpenConns大于0但小于新的MaxIdleConns，那么新的MaxIdleConns将被减少以匹配MaxOpenConns的限制。
//
//			n <= 0表示不保留空闲连接。
//
//			当前默认的最大空闲连接数是2。在未来的版本中，这可能会改变。
//		*/
//		sqlDB.SetMaxIdleConns(maxIdleConns)
//		/*
//			SetMaxOpenConns设置数据库的最大打开连接数。
//
//			如果MaxIdleConns大于0并且新的MaxIdleConns小于MaxIdleConns，那么MaxIdleConns将会减少以匹配新的MaxOpenConns限制。
//
//			如果n <= 0，则对打开连接的数量没有限制。默认值是0(无限)。
//		*/
//		sqlDB.SetMaxOpenConns(maxOpenConns)
//		/*
//			设置一个连接可以被重用的最大时间。
//
//			过期的连接可能在重用之前被惰性关闭。
//
//			如果d <= 0，连接不会因为连接的老化而关闭。
//		*/
//		sqlDB.SetConnMaxLifetime(connMaxLifetime)
//	}
//	db = gormDB
//	return err
//}
//
//// NewDB 外部获取*gorm.DB实例的方式
//func NewDB() *gorm.DB {
//	return db
//}
