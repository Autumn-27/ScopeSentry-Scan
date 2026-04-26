// configupdater-------------------------------------
// @file      : webfinger.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2025/12/29 20:58
// -------------------------------------------

package configupdater

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Autumn-27/ScopeSentry-Scan/internal/types"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"

	goahocorasick "github.com/anknown/ahocorasick"
	"gopkg.in/yaml.v3"
)

// NewACMatcher 创建新的AC自动机匹配器
func NewACMatcher() *types.ACMatcher {
	return &types.ACMatcher{
		TitlePatterns:     make([]types.PatternInfo, 0),
		HeaderPatterns:    make([]types.PatternInfo, 0),
		BodyPatterns:      make([]types.PatternInfo, 0),
		TitlePatternMap:   make(map[string]int),
		HeaderPatternMap:  make(map[string]int),
		BodyPatternMap:    make(map[string]int),
		FingerprintMap:    make(map[string]*types.Fingerprint),
		NonACFingerprints: make([]*types.Fingerprint, 0),
	}
}

// LoadFingerprintsFromDir 从目录加载所有指纹文件
func LoadFingerprintsFromDir(dirPath string) ([]*types.Fingerprint, error) {
	var fingerprints []*types.Fingerprint

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".yaml") {
			fp, err := LoadFingerprintWithID(path)
			if err != nil {
				// 忽略加载失败的文件，继续处理其他文件
				return nil
			}
			fingerprints = append(fingerprints, fp)
		}

		return nil
	})

	return fingerprints, err
}

// LoadFingerprintWithID 加载带ID的指纹文件
func LoadFingerprintWithID(filepath string) (*types.Fingerprint, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read fingerprint file: %w", err)
	}

	var fingerprint types.Fingerprint
	if err := yaml.Unmarshal(data, &fingerprint); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &fingerprint, nil
}

// extractPatternsFromRule 从rule中提取patterns
// 返回提取到的pattern信息列表，如果无法提取则返回空列表
// 对于OR逻辑的条件组，会收集组内所有符合条件的pattern
// 对于AND逻辑的rule，如果包含OR组，保留OR组内的所有patterns；否则选择最佳的pattern
func extractPatternsFromRule(rule types.Rule, fingerprint *types.Fingerprint, ruleIndex int) []types.PatternInfo {
	// 从rule的conditions中提取patterns
	patterns := extractPatternsFromConditionGroup(rule.Conditions, rule.Logic, fingerprint, ruleIndex)

	if len(patterns) == 0 {
		return nil
	}

	return patterns
}

func isConditionGroup(condition types.Condition) bool {
	return condition.Logic != "" && condition.Location == ""
}

// extractPatternsFromConditionGroup 从条件组中提取patterns（递归处理无限嵌套）
// logic: 当前条件组的逻辑（AND或OR）
func extractPatternsFromConditionGroup(conditions []types.Condition, logic string, fingerprint *types.Fingerprint, ruleIndex int) []types.PatternInfo {
	var allPatterns []types.PatternInfo

	// 遍历所有条件
	for _, condition := range conditions {
		// 如果是嵌套条件组
		if isConditionGroup(condition) {
			// 递归提取嵌套组中的patterns
			nestedPatterns := extractPatternsFromConditionGroup(condition.Conditions, condition.Logic, fingerprint, ruleIndex)
			allPatterns = append(allPatterns, nestedPatterns...)
			continue
		}

		// 处理普通条件：只处理contains类型，且location为title/header/body
		if condition.MatchType == "contains" &&
			(condition.Location == "title" || condition.Location == "header" || condition.Location == "body") &&
			condition.Pattern != "" {
			allPatterns = append(allPatterns, types.PatternInfo{
				Pattern:       condition.Pattern,
				Location:      condition.Location,
				FingerprintID: fingerprint.ID,
				RuleIndex:     ruleIndex,
			})
		}
	}

	if len(allPatterns) == 0 {
		return nil
	}

	// 根据逻辑类型处理patterns
	if logic == "OR" {
		// OR逻辑：返回所有patterns（因为OR组内的所有patterns都需要匹配）
		// 去重：相同的pattern只保留一个（基于pattern字符串和location）
		return deduplicatePatterns(allPatterns)
	} else {
		// AND逻辑：需要检查是否包含来自OR组的patterns
		// 如果所有patterns都来自同一个OR组（通过递归收集），保留所有
		// 否则只选择一个最佳的pattern

		// 检查是否有OR组（通过检查是否有多个相同优先级的patterns）
		// 更简单的方法：如果patterns来自不同的条件组层级，可能需要选择最佳
		// 但实际上，如果AND组内包含OR组，OR组的所有patterns都应该保留
		// 如果AND组内只有AND子组或普通条件，只选择一个最佳的

		// 简化处理：对于AND逻辑，如果patterns数量大于1，检查是否来自OR组
		// 由于我们无法直接判断patterns的来源，采用保守策略：
		// 如果AND组内直接或间接包含OR组，应该保留所有patterns
		// 否则只选择一个最佳的

		// 检查conditions中是否包含OR组（递归检查）
		hasORGroup := hasORGroupInConditions(conditions)

		if hasORGroup {
			// 如果包含OR组，保留所有patterns（因为OR组内的所有patterns都需要匹配）
			return allPatterns
		} else {
			// 否则只选择一个最佳的pattern
			best := selectBestPattern(allPatterns)
			if best != nil {
				return []types.PatternInfo{*best}
			}
			return nil
		}
	}
}

// hasORGroupInConditions 递归检查conditions中是否包含OR逻辑的条件组
func hasORGroupInConditions(conditions []types.Condition) bool {
	for _, condition := range conditions {
		if isConditionGroup(condition) {
			if condition.Logic == "OR" {
				return true
			}
			// 递归检查嵌套组
			if hasORGroupInConditions(condition.Conditions) {
				return true
			}
		}
	}
	return false
}

// hasRequestConditionInConditions 递归检查条件中是否包含request条件
func hasRequestConditionInConditions(conditions []types.Condition) bool {
	for _, condition := range conditions {
		if isConditionGroup(condition) {
			if hasRequestConditionInConditions(condition.Conditions) {
				return true
			}
			continue
		}
		if condition.Location == "request" {
			return true
		}
	}
	return false
}

// deduplicatePatterns 去重patterns（基于pattern字符串和location）
func deduplicatePatterns(patterns []types.PatternInfo) []types.PatternInfo {
	if len(patterns) <= 1 {
		return patterns
	}

	seen := make(map[string]bool)
	result := make([]types.PatternInfo, 0, len(patterns))

	for _, p := range patterns {
		key := p.Location + ":" + p.Pattern
		if !seen[key] {
			seen[key] = true
			result = append(result, p)
		}
	}

	return result
}

// selectBestPattern 根据优先级和长度选择最佳pattern（用于AND逻辑）
// 优先级：title > header > body
// 对于相同优先级，选择最长的
func selectBestPattern(candidates []types.PatternInfo) *types.PatternInfo {
	if len(candidates) == 0 {
		return nil
	}

	// 定义优先级
	priority := map[string]int{
		"title":  3,
		"header": 2,
		"body":   1,
	}

	var best *types.PatternInfo
	bestPriority := 0
	bestLength := 0

	for i := range candidates {
		candidate := &candidates[i]
		p := priority[candidate.Location]
		length := len(candidate.Pattern)

		// 优先级更高，或者优先级相同但长度更长
		if p > bestPriority || (p == bestPriority && length > bestLength) {
			best = candidate
			bestPriority = p
			bestLength = length
		}
	}

	return best
}

// BuildACMatcher 构建AC自动机匹配器
func BuildACMatcher(fingerprints []*types.Fingerprint) *types.ACMatcher {
	matcher := NewACMatcher()

	// 使用map进行全局去重，key为 location:pattern，value为PatternInfo
	// 这样可以避免重复的patterns被添加到AC自动机中（不影响匹配准确性）
	titlePatternMap := make(map[string]types.PatternInfo)
	headerPatternMap := make(map[string]types.PatternInfo)
	bodyPatternMap := make(map[string]types.PatternInfo)

	// 遍历所有fingerprint，建立FingerprintMap（优化内存，避免在PatternInfo中重复存储）
	for _, fingerprint := range fingerprints {
		matcher.FingerprintMap[fingerprint.ID] = fingerprint
	}

	// 遍历所有fingerprint
	for _, fingerprint := range fingerprints {
		// request 条件依赖主动请求，不能仅通过 title/header/body 的 AC 预筛可靠覆盖
		// 一旦存在 request 条件，整条指纹降级为非 AC 流程，避免漏报
		hasRequestCondition := false
		for _, rule := range fingerprint.Rules {
			if hasRequestConditionInConditions(rule.Conditions) {
				hasRequestCondition = true
				break
			}
		}
		if hasRequestCondition {
			matcher.NonACFingerprints = append(matcher.NonACFingerprints, fingerprint)
			continue
		}

		// 先检查所有rules是否都能提取到pattern
		allRulesHavePattern := true
		allRulePatterns := make([][]types.PatternInfo, 0, len(fingerprint.Rules))

		// 遍历每个rule，尝试提取patterns
		for ruleIndex, rule := range fingerprint.Rules {
			patterns := extractPatternsFromRule(rule, fingerprint, ruleIndex)
			if len(patterns) == 0 {
				// 如果任何一个rule无法提取pattern，标记为false
				allRulesHavePattern = false
				// 不立即break，继续检查其他rules以便调试
				break
			} else {
				allRulePatterns = append(allRulePatterns, patterns)
			}
		}

		// 如果所有rules都能提取到pattern，才添加到AC自动机
		if !allRulesHavePattern {
			// 如果有任何一个rule无法提取pattern，整个fingerprint放入非AC列表
			matcher.NonACFingerprints = append(matcher.NonACFingerprints, fingerprint)
			continue
		}

		// 如果所有rules都能提取到pattern，才添加到AC自动机
		if allRulesHavePattern && len(allRulePatterns) > 0 {
			// 将所有rule的patterns添加到对应的map中（仅去重，不进行长度过滤，确保不漏报）
			for _, rulePatterns := range allRulePatterns {
				for _, patternInfo := range rulePatterns {
					key := patternInfo.Location + ":" + patternInfo.Pattern

					switch patternInfo.Location {
					case "title":
						// 全局去重：如果已存在，保留第一个（相同pattern只保留一次，不影响匹配）
						if _, exists := titlePatternMap[key]; !exists {
							titlePatternMap[key] = patternInfo
						}
					case "header":
						if _, exists := headerPatternMap[key]; !exists {
							headerPatternMap[key] = patternInfo
						}
					case "body":
						if _, exists := bodyPatternMap[key]; !exists {
							bodyPatternMap[key] = patternInfo
						}
					}
				}
			}
		} else {
			// 如果有任何一个rule无法提取pattern，整个fingerprint放入非AC列表
			matcher.NonACFingerprints = append(matcher.NonACFingerprints, fingerprint)
		}
	}

	// 将map转换为slice，并构建patterns列表用于AC自动机
	titlePatterns := make([]string, 0, len(titlePatternMap))
	headerPatterns := make([]string, 0, len(headerPatternMap))
	bodyPatterns := make([]string, 0, len(bodyPatternMap))

	for _, patternInfo := range titlePatternMap {
		matcher.TitlePatterns = append(matcher.TitlePatterns, patternInfo)
		titlePatterns = append(titlePatterns, patternInfo.Pattern)
	}

	for _, patternInfo := range headerPatternMap {
		matcher.HeaderPatterns = append(matcher.HeaderPatterns, patternInfo)
		headerPatterns = append(headerPatterns, patternInfo.Pattern)
	}

	for _, patternInfo := range bodyPatternMap {
		matcher.BodyPatterns = append(matcher.BodyPatterns, patternInfo)
		bodyPatterns = append(bodyPatterns, patternInfo.Pattern)
	}

	// 清理临时map
	titlePatternMap = nil
	headerPatternMap = nil
	bodyPatternMap = nil

	// 构建AC自动机（使用Double-Array Trie实现，内存占用极低）

	// 将string patterns转换为[][]rune格式，并建立pattern到索引的映射
	if len(titlePatterns) > 0 {
		// 建立pattern到索引的映射
		for i, pattern := range titlePatterns {
			matcher.TitlePatternMap[pattern] = i
		}

		// 转换为[][]rune
		titlePatternsRune := make([][]rune, len(titlePatterns))
		for i, pattern := range titlePatterns {
			titlePatternsRune[i] = []rune(pattern)
		}

		// 构建Double-Array Trie AC自动机
		matcher.TitleMatcher = &goahocorasick.Machine{}
		if err := matcher.TitleMatcher.Build(titlePatternsRune); err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("构建Title AC自动机失败: %v", err))
		}

		titlePatternsRune = nil
	}

	if len(headerPatterns) > 0 {
		// 建立pattern到索引的映射
		for i, pattern := range headerPatterns {
			matcher.HeaderPatternMap[pattern] = i
		}

		// 转换为[][]rune
		headerPatternsRune := make([][]rune, len(headerPatterns))
		for i, pattern := range headerPatterns {
			headerPatternsRune[i] = []rune(pattern)
		}

		// 构建Double-Array Trie AC自动机
		matcher.HeaderMatcher = &goahocorasick.Machine{}
		if err := matcher.HeaderMatcher.Build(headerPatternsRune); err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("构建Header AC自动机失败: %v", err))
		}

		headerPatternsRune = nil
	}

	if len(bodyPatterns) > 0 {
		// 建立pattern到索引的映射
		for i, pattern := range bodyPatterns {
			matcher.BodyPatternMap[pattern] = i
		}

		// 转换为[][]rune
		bodyPatternsRune := make([][]rune, len(bodyPatterns))
		for i, pattern := range bodyPatterns {
			bodyPatternsRune[i] = []rune(pattern)
		}

		// 构建Double-Array Trie AC自动机
		matcher.BodyMatcher = &goahocorasick.Machine{}
		if err := matcher.BodyMatcher.Build(bodyPatternsRune); err != nil {
			logger.SlogErrorLocal(fmt.Sprintf("构建Body AC自动机失败: %v", err))
		}

		bodyPatternsRune = nil
	}

	// 清理临时数据
	titlePatterns = nil
	headerPatterns = nil
	bodyPatterns = nil

	return matcher
}

//
//func main() {
//	//test_main()
//	// 从fingers文件夹加载所有指纹文件
//	fmt.Println("开始加载指纹文件...")
//	fingerprints, err := LoadFingerprintsFromDir("fingers")
//	if err != nil {
//		log.Fatalf("加载指纹文件失败: %v", err)
//	}
//	fmt.Printf("✓ 成功加载 %d 个指纹文件\n", len(fingerprints))
//
//	// 构建AC自动机
//	fmt.Println("\n开始构建AC自动机...")
//	matcher := BuildACMatcher(fingerprints)
//	//
//	//// 统计信息
//	//totalFingerprints := len(fingerprints)
//	//acFingerprints := totalFingerprints - len(matcher.NonACFingerprints)
//	//nonACFingerprints := len(matcher.NonACFingerprints)
//	//
//	//// 输出统计结果
//	//fmt.Println("\n" + strings.Repeat("=", 60))
//	//fmt.Println("AC自动机构建完成 - 统计结果")
//	//fmt.Println(strings.Repeat("=", 60))
//	//fmt.Printf("总指纹数量:        %d\n", totalFingerprints)
//	//fmt.Printf("可以使用AC的指纹:   %d (%.2f%%)\n", acFingerprints, float64(acFingerprints)/float64(totalFingerprints)*100)
//	//fmt.Printf("无法使用AC的指纹:   %d (%.2f%%)\n", nonACFingerprints, float64(nonACFingerprints)/float64(totalFingerprints)*100)
//	//fmt.Println(strings.Repeat("=", 60))
//	//
//	//// Pattern统计
//	//fmt.Printf("\nPattern统计:\n")
//	//fmt.Printf("  Title patterns:  %d\n", len(matcher.TitlePatterns))
//	//fmt.Printf("  Header patterns: %d\n", len(matcher.HeaderPatterns))
//	//fmt.Printf("  Body patterns:   %d\n", len(matcher.BodyPatterns))
//	//fmt.Printf("  总patterns:      %d\n", len(matcher.TitlePatterns)+len(matcher.HeaderPatterns)+len(matcher.BodyPatterns))
//
//	// 如果需要，可以输出无法使用AC的指纹列表
//	//if nonACFingerprints > 0 && nonACFingerprints <= 20 {
//	//	fmt.Printf("\n无法使用AC的指纹列表 (前%d个):\n", nonACFingerprints)
//	//	for i, fp := range matcher.NonACFingerprints {
//	//		if i >= 20 {
//	//			break
//	//		}
//	//		fmt.Printf("  [%d] %s (ID: %s)\n", i+1, fp.Name, fp.ID)
//	//	}
//	//} else if nonACFingerprints > 20 {
//	//	fmt.Printf("\n无法使用AC的指纹列表 (前20个):\n")
//	//	for i := 0; i < 20 && i < len(matcher.NonACFingerprints); i++ {
//	//		fp := matcher.NonACFingerprints[i]
//	//		fmt.Printf("  [%d] %s (ID: %s)\n", i+1, fp.Name, fp.ID)
//	//	}
//	//	fmt.Printf("  ... 还有 %d 个指纹无法使用AC\n", nonACFingerprints-20)
//	//}
//
//	// 匹配测试
//	fmt.Println("\n" + strings.Repeat("=", 60))
//	fmt.Println("AC自动机匹配测试")
//	fmt.Println(strings.Repeat("=", 60))
//
//	// 测试数据（可以修改这些值进行测试）
//	testTitle := "D-Link DSL-2640B 下一代防火墙安全网关"
//	testHeader := "Server: mini_httpd\nContent-Type: text/html Server: AvigilonGateway"
//	testBody := "Product : DSL-2640B"
//
//	fmt.Printf("测试数据:\n")
//	fmt.Printf("  Title:  %q\n", testTitle)
//	fmt.Printf("  Header: %q\n", testHeader)
//	fmt.Printf("  Body:   %q\n", testBody)
//
//	// 收集命中的指纹（按ID去重）
//	type MatchedFingerprint struct {
//		FingerprintID   string
//		FingerprintName string
//		MatchedRules    []int // 匹配到的rule索引列表
//	}
//
//	matchedFingerprintsMap := make(map[string]*MatchedFingerprint) // key: fingerprintID
//
//	// 匹配title并收集指纹
//	if matcher.TitleMatcher != nil {
//		titleMatches := matcher.TitleMatcher.Match([]byte(testTitle))
//		for _, patternIndex := range titleMatches {
//			if patternIndex < len(matcher.TitlePatterns) {
//				p := matcher.TitlePatterns[patternIndex]
//				if fp, exists := matchedFingerprintsMap[p.Fingerprint.ID]; exists {
//					// 检查ruleIndex是否已存在
//					ruleExists := false
//					for _, ruleIdx := range fp.MatchedRules {
//						if ruleIdx == p.RuleIndex {
//							ruleExists = true
//							break
//						}
//					}
//					if !ruleExists {
//						fp.MatchedRules = append(fp.MatchedRules, p.RuleIndex)
//					}
//				} else {
//					matchedFingerprintsMap[p.Fingerprint.ID] = &MatchedFingerprint{
//						FingerprintID:   p.Fingerprint.ID,
//						FingerprintName: p.Fingerprint.Name,
//						MatchedRules:    []int{p.RuleIndex},
//					}
//				}
//			}
//		}
//	}
//
//	// 匹配header并收集指纹
//	if matcher.HeaderMatcher != nil {
//		headerMatches := matcher.HeaderMatcher.Match([]byte(testHeader))
//		for _, patternIndex := range headerMatches {
//			if patternIndex < len(matcher.HeaderPatterns) {
//				p := matcher.HeaderPatterns[patternIndex]
//				if fp, exists := matchedFingerprintsMap[p.Fingerprint.ID]; exists {
//					// 检查ruleIndex是否已存在
//					ruleExists := false
//					for _, ruleIdx := range fp.MatchedRules {
//						if ruleIdx == p.RuleIndex {
//							ruleExists = true
//							break
//						}
//					}
//					if !ruleExists {
//						fp.MatchedRules = append(fp.MatchedRules, p.RuleIndex)
//					}
//				} else {
//					matchedFingerprintsMap[p.Fingerprint.ID] = &MatchedFingerprint{
//						FingerprintID:   p.Fingerprint.ID,
//						FingerprintName: p.Fingerprint.Name,
//						MatchedRules:    []int{p.RuleIndex},
//					}
//				}
//			}
//		}
//	}
//
//	// 匹配body并收集指纹
//	if matcher.BodyMatcher != nil {
//		bodyMatches := matcher.BodyMatcher.Match([]byte(testBody))
//		for _, patternIndex := range bodyMatches {
//			if patternIndex < len(matcher.BodyPatterns) {
//				p := matcher.BodyPatterns[patternIndex]
//				if fp, exists := matchedFingerprintsMap[p.Fingerprint.ID]; exists {
//					// 检查ruleIndex是否已存在
//					ruleExists := false
//					for _, ruleIdx := range fp.MatchedRules {
//						if ruleIdx == p.RuleIndex {
//							ruleExists = true
//							break
//						}
//					}
//					if !ruleExists {
//						fp.MatchedRules = append(fp.MatchedRules, p.RuleIndex)
//					}
//				} else {
//					matchedFingerprintsMap[p.Fingerprint.ID] = &MatchedFingerprint{
//						FingerprintID:   p.Fingerprint.ID,
//						FingerprintName: p.Fingerprint.Name,
//						MatchedRules:    []int{p.RuleIndex},
//					}
//				}
//			}
//		}
//	}
//
//	// 输出匹配结果
//	fmt.Printf("\n匹配结果:\n")
//	fmt.Printf("  命中的指纹数量: %d (按ID去重后)\n", len(matchedFingerprintsMap))
//
//	if len(matchedFingerprintsMap) > 0 {
//		fmt.Println("\n命中的指纹ID列表:")
//		// 只输出去重后的ID
//		for _, fp := range matchedFingerprintsMap {
//			fmt.Printf("  %s\n", fp.FingerprintID)
//		}
//	} else {
//		fmt.Println("  未匹配到任何指纹")
//	}
//}
