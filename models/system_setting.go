package models

// GeneralConfig 通用配置结构体
type GeneralConfig struct {
	MaxThumbSize    uint `mapstructure:"max_thumb_size"`
	RegisterEnabled bool `mapstructure:"register_enabled"`
	MaxTags         int  `mapstructure:"max_tags"`
}

// MailConfig 邮件配置结构体
type MailConfig struct {
	Enabled       bool       `mapstructure:"enabled"`
	ServerAddress string     `mapstructure:"server_address"`
	SMTP          SMTPConfig `mapstructure:"smtp"`
}

// SMTPConfig SMTP配置结构体
type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	FromName string `mapstructure:"from_name"`
}

type GalleryConfig struct {
	Title           string `mapstructure:"title"`
	BackgroundImage string `mapstructure:"background_image"`
	CustomContent   string `mapstructure:"custom_content"`
}

// SystemSettings 系统设置结构体
type SystemSettings struct {
	General GeneralConfig `mapstructure:"general"`
	Mail    MailConfig    `mapstructure:"mail"`
	Gallery GalleryConfig `mapstructure:"gallery"`
}

// SystemSetting 系统设置模型
type SystemSetting struct {
	BaseModel
	Value SystemSettings `gorm:"type:json;serializer:json" json:"value"`
}
