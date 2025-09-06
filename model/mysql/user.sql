CREATE TABLE user (
    id BIGINT UNSIGNED AUTO_INCREMENT COMMENT '自增主键，内部关联使用',
    public_id VARCHAR(36) NOT NULL COMMENT '对外暴露的UserID，避免ID枚举攻击',
    nickname VARCHAR(100) DEFAULT NULL COMMENT '昵称',
    username VARCHAR(100) NOT NULL COMMENT '用户名',
    email VARCHAR(255) NOT NULL COMMENT '邮箱',
    email_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT '邮箱是否已验证 (0-未验证, 1-已验证)',
    phone VARCHAR(20) DEFAULT NULL COMMENT '手机号 (国际格式)',
    phone_verified TINYINT(1) NOT NULL DEFAULT 0 COMMENT '手机号是否已验证 (0-未验证, 1-已验证)',
    password_hash VARCHAR(255) NOT NULL COMMENT '加密后的密码',
    password_salt VARCHAR(255) DEFAULT NULL COMMENT '密码盐',
    mfa_secret VARCHAR(255) DEFAULT NULL COMMENT 'MFA秘钥 (加密存储)',
    mfa_enabled TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否启用了MFA (0-未启用, 1-启用)',
    account_status TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '账户状态 (1-正常, 2-锁定, 3-禁用)',
    failed_login_attempts TINYINT UNSIGNED NOT NULL DEFAULT 0 COMMENT '连续失败登录次数',
    lockout_until DATETIME DEFAULT NULL COMMENT '账户锁定截止时间',
    last_login_at DATETIME DEFAULT NULL COMMENT '最后一次登录时间',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted_at DATETIME DEFAULT NULL COMMENT '软删除时间',

    -- 主键
    PRIMARY KEY (id),

    -- 唯一约束
    UNIQUE KEY uq_user_public_id (public_id),
    UNIQUE KEY uq_user_username (username),
    UNIQUE KEY uq_user_email (email),

    -- 查询索引
    KEY idx_user_phone (phone),
    KEY idx_user_account_status (account_status),
    KEY idx_user_lockout (lockout_until),
    KEY idx_user_created_at (created_at),
    KEY idx_user_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';