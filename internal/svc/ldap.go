package svc

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"auth-service/internal/config"

	"github.com/go-ldap/ldap/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// LDAPProvider LDAP 认证提供者
type LDAPProvider struct {
	config config.LDAPConfig
}

// LDAPUserInfo LDAP 用户信息
type LDAPUserInfo struct {
	DN          string
	Username    string
	Email       string
	DisplayName string
	FirstName   string
	LastName    string
	Phone       string
	Groups      []string
	Attributes  map[string][]string
}

// NewLDAPProvider 创建 LDAP 提供者
func NewLDAPProvider(cfg config.LDAPConfig) (*LDAPProvider, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	provider := &LDAPProvider{
		config: cfg,
	}

	// 测试连接
	conn, err := provider.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP server: %w", err)
	}
	defer conn.Close()

	logx.Infof("LDAP provider initialized, server: %s:%d", cfg.Host, cfg.Port)
	return provider, nil
}

// connect 建立 LDAP 连接
func (p *LDAPProvider) connect() (*ldap.Conn, error) {
	var (
		conn *ldap.Conn
		err  error
	)

	address := fmt.Sprintf("%s:%d", p.config.Host, p.config.Port)

	if p.config.UseSSL {
		// LDAPS 连接
		tlsConfig := &tls.Config{
			InsecureSkipVerify: p.config.InsecureSkipTLS,
			MinVersion:         tls.VersionTLS12,
		}
		conn, err = ldap.DialTLS("tcp", address, tlsConfig)
	} else {
		// 普通连接
		conn, err = ldap.Dial("tcp", address)

		// 如果需要 StartTLS
		if err == nil && p.config.UseTLS {
			tlsConfig := &tls.Config{
				InsecureSkipVerify: p.config.InsecureSkipTLS,
				MinVersion:         tls.VersionTLS12,
				ServerName:         p.config.Host,
			}
			err = conn.StartTLS(tlsConfig)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to dial LDAP: %w", err)
	}

	// 设置超时
	conn.SetTimeout(30 * time.Second)

	return conn, nil
}

// bindAsAdmin 使用管理员账号绑定
func (p *LDAPProvider) bindAsAdmin(conn *ldap.Conn) error {
	if p.config.BindDN == "" {
		// 匿名绑定
		return conn.UnauthenticatedBind("")
	}
	return conn.Bind(p.config.BindDN, p.config.BindPassword)
}

// Authenticate 验证用户凭据
func (p *LDAPProvider) Authenticate(ctx context.Context, username, password string) (*LDAPUserInfo, error) {
	// 建立连接
	conn, err := p.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// 先用管理员账号绑定来搜索用户
	if err := p.bindAsAdmin(conn); err != nil {
		return nil, fmt.Errorf("failed to bind as admin: %w", err)
	}

	// 搜索用户
	userDN, userInfo, err := p.searchUser(conn, username)
	if err != nil {
		return nil, err
	}

	// 使用用户的 DN 和密码重新绑定来验证密码
	if err := conn.Bind(userDN, password); err != nil {
		ldapErr, ok := err.(*ldap.Error)
		if ok && ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials {
			return nil, fmt.Errorf("invalid credentials")
		}
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	logx.Infof("LDAP authentication successful for user: %s", username)
	return userInfo, nil
}

// searchUser 搜索用户
func (p *LDAPProvider) searchUser(conn *ldap.Conn, username string) (string, *LDAPUserInfo, error) {
	// 构建用户过滤器
	filter := strings.Replace(p.config.UserFilter, "%s", ldap.EscapeFilter(username), -1)

	// 构建属性列表
	attributes := p.config.UserAttributes
	if len(attributes) == 0 {
		attributes = []string{"*"} // 获取所有属性
	}

	searchRequest := ldap.NewSearchRequest(
		p.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		1,  // 只需要一个结果
		30, // 超时 30 秒
		false,
		filter,
		attributes,
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return "", nil, fmt.Errorf("failed to search user: %w", err)
	}

	if len(result.Entries) == 0 {
		return "", nil, fmt.Errorf("user not found: %s", username)
	}

	if len(result.Entries) > 1 {
		return "", nil, fmt.Errorf("multiple users found for: %s", username)
	}

	entry := result.Entries[0]
	userInfo := p.entryToUserInfo(entry)

	return entry.DN, userInfo, nil
}

// entryToUserInfo 将 LDAP 条目转换为用户信息
func (p *LDAPProvider) entryToUserInfo(entry *ldap.Entry) *LDAPUserInfo {
	userInfo := &LDAPUserInfo{
		DN:         entry.DN,
		Attributes: make(map[string][]string),
	}

	// 获取所有属性
	for _, attr := range entry.Attributes {
		userInfo.Attributes[attr.Name] = attr.Values
	}

	// 映射配置的属性
	if p.config.UsernameAttr != "" {
		userInfo.Username = entry.GetAttributeValue(p.config.UsernameAttr)
	}
	if p.config.EmailAttr != "" {
		userInfo.Email = entry.GetAttributeValue(p.config.EmailAttr)
	}
	if p.config.DisplayNameAttr != "" {
		userInfo.DisplayName = entry.GetAttributeValue(p.config.DisplayNameAttr)
	}

	// 尝试获取常见属性
	if userInfo.Username == "" {
		userInfo.Username = entry.GetAttributeValue("uid")
		if userInfo.Username == "" {
			userInfo.Username = entry.GetAttributeValue("sAMAccountName")
		}
		if userInfo.Username == "" {
			userInfo.Username = entry.GetAttributeValue("cn")
		}
	}
	if userInfo.Email == "" {
		userInfo.Email = entry.GetAttributeValue("mail")
	}
	if userInfo.DisplayName == "" {
		userInfo.DisplayName = entry.GetAttributeValue("displayName")
		if userInfo.DisplayName == "" {
			userInfo.DisplayName = entry.GetAttributeValue("cn")
		}
	}

	// 获取名字
	userInfo.FirstName = entry.GetAttributeValue("givenName")
	userInfo.LastName = entry.GetAttributeValue("sn")
	userInfo.Phone = entry.GetAttributeValue("telephoneNumber")

	// 获取组成员资格
	if p.config.GroupMemberAttr != "" {
		userInfo.Groups = entry.GetAttributeValues(p.config.GroupMemberAttr)
	} else {
		userInfo.Groups = entry.GetAttributeValues("memberOf")
	}

	return userInfo
}

// GetUserGroups 获取用户的组
func (p *LDAPProvider) GetUserGroups(ctx context.Context, userDN string) ([]string, error) {
	conn, err := p.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	if err := p.bindAsAdmin(conn); err != nil {
		return nil, fmt.Errorf("failed to bind as admin: %w", err)
	}

	// 如果配置了组过滤器，使用它来搜索组
	if p.config.GroupFilter != "" {
		filter := strings.Replace(p.config.GroupFilter, "%s", ldap.EscapeFilter(userDN), -1)

		searchRequest := ldap.NewSearchRequest(
			p.config.BaseDN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0,
			30,
			false,
			filter,
			[]string{"cn", "dn"},
			nil,
		)

		result, err := conn.Search(searchRequest)
		if err != nil {
			return nil, fmt.Errorf("failed to search groups: %w", err)
		}

		groups := make([]string, len(result.Entries))
		for i, entry := range result.Entries {
			groups[i] = entry.GetAttributeValue("cn")
		}
		return groups, nil
	}

	// 否则从用户条目中获取 memberOf 属性
	searchRequest := ldap.NewSearchRequest(
		userDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		1,
		30,
		false,
		"(objectClass=*)",
		[]string{"memberOf"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}

	if len(result.Entries) == 0 {
		return []string{}, nil
	}

	return result.Entries[0].GetAttributeValues("memberOf"), nil
}

// SearchUsers 搜索用户列表
func (p *LDAPProvider) SearchUsers(ctx context.Context, filter string, limit int) ([]*LDAPUserInfo, error) {
	conn, err := p.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	if err := p.bindAsAdmin(conn); err != nil {
		return nil, fmt.Errorf("failed to bind as admin: %w", err)
	}

	attributes := p.config.UserAttributes
	if len(attributes) == 0 {
		attributes = []string{"*"}
	}

	searchRequest := ldap.NewSearchRequest(
		p.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		limit,
		30,
		false,
		filter,
		attributes,
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	users := make([]*LDAPUserInfo, len(result.Entries))
	for i, entry := range result.Entries {
		users[i] = p.entryToUserInfo(entry)
	}

	return users, nil
}

// TestConnection 测试连接
func (p *LDAPProvider) TestConnection(ctx context.Context) error {
	conn, err := p.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	return p.bindAsAdmin(conn)
}

// IsEnabled 检查是否启用
func (p *LDAPProvider) IsEnabled() bool {
	return p != nil && p.config.Enabled
}

// GetConfig 获取配置（用于调试）
func (p *LDAPProvider) GetConfig() config.LDAPConfig {
	// 返回配置副本，隐藏敏感信息
	cfg := p.config
	cfg.BindPassword = "***"
	return cfg
}
