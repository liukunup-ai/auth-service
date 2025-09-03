CREATE TABLE user (
    id BIGINT UNSIGNED AUTO_INCREMENT, -- 自增主键，内部关联使用，性能好
    public_id VARCHAR(36) NOT NULL, -- 对外暴露的用户UUID，避免ID枚举攻击，用于API返回
    username VARCHAR(255) NOT NULL, -- 用户名
    email VARCHAR(255) NOT NULL, -- 邮箱
    email_verified TINYINT(1) NOT NULL DEFAULT 0, -- 邮箱是否已验证 (使用0/1而非BOOL为兼容性)
    phone VARCHAR(20) NULL, -- 手机号 (国际格式)
    phone_verified TINYINT(1) NOT NULL DEFAULT 0, -- 手机号是否已验证
    password_hash VARCHAR(255) NOT NULL, -- 加密后的密码 (如 bcrypt)
    password_salt VARCHAR(255) NULL, -- 密码盐 (如果算法不需要则可为空)
    mfa_secret VARCHAR(255) NULL, -- MFA秘钥 (加密存储)
    mfa_enabled TINYINT(1) NOT NULL DEFAULT 0, -- 是否启用了MFA
    account_locked TINYINT(1) NOT NULL DEFAULT 0, -- 账户是否被锁定
    failed_login_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0, -- 连续失败登录次数
    lockout_until DATETIME NULL, -- 账户锁定直到何时
    last_login_at DATETIME NULL, -- 最后一次登录时间
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP, -- 更新时间
    -- 唯一约束
    UNIQUE INDEX uq_user_public_id (public_id),
    UNIQUE INDEX uq_user_username (username),
    UNIQUE INDEX uq_user_email (email),
    -- 外键和查询索引
    INDEX idx_user_phone (phone),
    INDEX idx_user_created_at (created_at),
    INDEX idx_user_locked (account_locked, lockout_until), -- 用于检查账户状态的查询
    -- 主键
    PRIMARY KEY (id)
) ENGINE=InnoDB CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci COMMENT '用户表';