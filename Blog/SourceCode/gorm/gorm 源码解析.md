# gorm 源码解析

```sh
docker run -d --name=mysql-server -p 3306:3306 -v mysql-data:/var/lib/mysql -e MYSQL_ROOT_PASSWORD=123456 mysql:5.7
```



## 创建db连接

```go
func GetDB() *gorm.DB {
	dsn := "root:123456@tcp(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return db
}
```

创建连接使用Open方法，创建db连接

```go
func Open(dialector Dialector, opts ...Option) (db *DB, err error) 
```

Open 包含2个参数：

- dialector：连接器
- opts：gorm的附带参数

Option 是一个接口

```go
// Option gorm option interface
type Option interface {
   Apply(*Config) error // 加载参数
   AfterInitialize(*DB) error // 回调方法
}
```





Dialector 是一个接口，表示GORM 数据库的方言

```go
// Dialector GORM database dialector
type Dialector interface {
   Name() string
   Initialize(*DB) error
   Migrator(db *DB) Migrator
   DataTypeOf(*schema.Field) string
   DefaultValueOf(*schema.Field) clause.Expression
   BindVarTo(writer clause.Writer, stmt *Statement, v interface{})
   QuoteTo(clause.Writer, string)
   Explain(sql string, vars ...interface{}) string
}
```



以mysql为例，展示创建Dialector的过程

```go
func Open(dsn string) gorm.Dialector {
	dsnConf, _ := mysql.ParseDSN(dsn)
	return &Dialector{Config: &Config{DSN: dsn, DSNConfig: dsnConf}}
}
```

这里的mysql包是`github.com/go-sql-driver/mysql` 是mysql的驱动包。

```go
// ParseDSN parses the DSN string to a Config
func ParseDSN(dsn string) (cfg *Config, err error) 
```

ParseDSN 将 dsn 连接地址转换为 dsn 配置  Config

```go
return &Dialector{Config: &Config{DSN: dsn, DSNConfig: dsnConf}}
```

然后将 dsn 配置和dsn地址 包装成 Dialector 对象

可以看出 Dialector 只是一种 数据库配置，并不包含实际的数据库连接。



接下来仔细看看Open方法

```go
// Open initialize db session based on dialector
func Open(dialector Dialector, opts ...Option) (db *DB, err error) {
	config := &Config{}

    // 参数排序，将Config排到Option前面
	sort.Slice(opts, func(i, j int) bool {
		_, isConfig := opts[i].(*Config)
		_, isConfig2 := opts[j].(*Config)
		return isConfig && !isConfig2
	})

	for _, opt := range opts {
		if opt != nil {
            // 加载参数
			if applyErr := opt.Apply(config); applyErr != nil {
				return nil, applyErr
			}
            // 执行回调函数
			defer func(opt Option) {
				if errr := opt.AfterInitialize(db); errr != nil {
					err = errr
				}
			}(opt)
		}
	}

    // dialector 实现了 Config接口，则加载参数
	if d, ok := dialector.(interface{ Apply(*Config) error }); ok {
		if err = d.Apply(config); err != nil {
			return
		}
	}
    
    // 以下都是默认值设置
	if config.NamingStrategy == nil {
		config.NamingStrategy = schema.NamingStrategy{}
	}

	if config.Logger == nil {
		config.Logger = logger.Default
	}

	if config.NowFunc == nil {
		config.NowFunc = func() time.Time { return time.Now().Local() }
	}

	if dialector != nil {
		config.Dialector = dialector
	}

	if config.Plugins == nil {
		config.Plugins = map[string]Plugin{}
	}

	if config.cacheStore == nil {
		config.cacheStore = &sync.Map{}
	}

    // 创建db对象
	db = &DB{Config: config, clone: 1}

    // 初始化回调
	db.callbacks = initializeCallbacks(db)

	if config.ClauseBuilders == nil {
		config.ClauseBuilders = map[string]clause.ClauseBuilder{}
	}

	if config.Dialector != nil {
		err = config.Dialector.Initialize(db)
	}

	preparedStmt := &PreparedStmtDB{
		ConnPool:    db.ConnPool,
		Stmts:       make(map[string]*Stmt),
		Mux:         &sync.RWMutex{},
		PreparedSQL: make([]string, 0, 100),
	}
	db.cacheStore.Store(preparedStmtDBKey, preparedStmt)

	if config.PrepareStmt {
		db.ConnPool = preparedStmt
	}

	db.Statement = &Statement{
		DB:       db,
		ConnPool: db.ConnPool,
		Context:  context.Background(),
		Clauses:  map[string]clause.Clause{},
	}

	if err == nil && !config.DisableAutomaticPing {
		if pinger, ok := db.ConnPool.(interface{ Ping() error }); ok {
			err = pinger.Ping()
		}
	}

	if err != nil {
		config.Logger.Error(context.Background(), "failed to initialize database, got error %v", err)
	}

	return
}

```





