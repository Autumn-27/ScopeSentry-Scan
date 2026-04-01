// types-------------------------------------
// @file      : finger.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/28 23:11
// -------------------------------------------

package types

import goahocorasick "github.com/anknown/ahocorasick"

// Fingerprint 指纹定义
type Fingerprint struct {
	Name           string   `yaml:"name"`
	ID             string   `yaml:"id"`
	Tags           []string `yaml:"tags"` // 关联 POC
	Category       string   `yaml:"category"`
	ParentCategory string   `yaml:"parent_category"`
	Company        string   `yaml:"company"`
	Rules          []Rule   `yaml:"rules"`
}

type FingerprintYaml struct {
	Fingerprint Fingerprint `yaml:"fingerprint"`
}

// Rule 规则定义
type Rule struct {
	Logic      string      `yaml:"logic"` // AND 或 OR
	Conditions []Condition `yaml:"conditions"`
}

// Condition 条件定义（支持普通条件和嵌套条件组）
type Condition struct {
	// 普通条件字段
	Location    string `yaml:"location,omitempty"`   // body, header, title, request
	MatchType   string `yaml:"match_type,omitempty"` // regex, contains, extract, active
	Pattern     string `yaml:"pattern,omitempty"`
	Group       int    `yaml:"group,omitempty"`
	SaveAs      string `yaml:"save_as,omitempty"`
	Path        string `yaml:"path,omitempty"`
	DynamicPath string `yaml:"dynamic_path,omitempty"`
	Method      string `yaml:"method,omitempty"`

	// 嵌套条件组字段
	Logic      string      `yaml:"logic,omitempty"`      // AND 或 OR（用于嵌套组）
	Conditions []Condition `yaml:"conditions,omitempty"` // 子条件或嵌套组
}

type PatternInfo struct {
	Pattern       string // pattern字符串
	Location      string // title, header, body
	FingerprintID string // 关联的fingerprint ID（优化内存，避免重复存储Fingerprint指针）
	RuleIndex     int    // 关联的rule索引
}

type WebFingerCore struct {
	ACMatcher *ACMatcher
}

type ACMatcher struct {
	TitleMatcher   *goahocorasick.Machine
	HeaderMatcher  *goahocorasick.Machine
	BodyMatcher    *goahocorasick.Machine
	TitlePatterns  []PatternInfo
	HeaderPatterns []PatternInfo
	BodyPatterns   []PatternInfo
	// Pattern到索引的映射（用于根据匹配结果找到PatternInfo）
	TitlePatternMap  map[string]int // pattern字符串 -> PatternInfo索引
	HeaderPatternMap map[string]int
	BodyPatternMap   map[string]int
	// Fingerprint ID到Fingerprint的映射（优化内存，避免在PatternInfo中重复存储）
	FingerprintMap map[string]*Fingerprint // fingerprintID -> Fingerprint
	// 无法使用AC自动机的fingerprint列表
	NonACFingerprints []*Fingerprint
}
