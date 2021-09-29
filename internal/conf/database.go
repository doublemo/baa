package conf

type (
	// DBMySQLConfig GORM数据库配置
	DBMySQLConfig struct {
		DNS string `alias:"dns"`

		// TablePrefix 表名前缀
		TablePrefix string `alias:"tablePrefix"`

		// SkipInitializeWithVersion 根据当前 MySQL 版本自动配置
		SkipInitializeWithVersion bool `alias:"skipInitializeWithVersion"`

		// DefaultStringSize string 类型字段的默认长度
		DefaultStringSize uint `alias:"defaultStringSize" default:"255"`

		// DefaultDatetimePrecision
		DefaultDatetimePrecision int `alias:"defaultDatetimePrecision" default:"2"`

		// DisableDatetimePrecision 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DisableDatetimePrecision bool `alias:"disableDatetimePrecision"`

		// DontSupportRenameIndex 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameIndex bool `alias:"dontSupportRenameIndex"`

		// DontSupportRenameColumn 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		DontSupportRenameColumn bool `alias:"dontSupportRenameColumn"`

		// DontSupportForShareClause
		DontSupportForShareClause bool `alias:"dontSupportForShareClause"`

		// Resolver 基本GORM多数据库配置
		Resolver []DBResolverConfig `alias:"resolver"`

		// ConnMaxIdleTime 连接最大空闲时间 / s
		ConnMaxIdleTime int `alias:"maxIdleTime" default:"3600"`

		// ConnMaxLifetime 连接最大生命周期
		ConnMaxLifetime int `alias:"maxLifetime" default:"7200"`

		// MaxIdleConns 最大空闲连接
		MaxIdleConns int `alias:"maxIdleConns" default:"100"`

		// MaxOpenConns 最大连接数
		MaxOpenConns int `alias:"maxOpenConns" default:"200"`
	}

	// DBResolverConfig 基本GORM多数据库配置
	DBResolverConfig struct {
		Sources  []string `alias:"sources"`
		Replicas []string `alias:"replicas"`
		Policy   string   `alias:"policy" default:"random"`
		Tables   []string `alias:"tables"`
	}
)
