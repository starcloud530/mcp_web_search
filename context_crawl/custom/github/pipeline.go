package github

import (
	"context_crawl/base/colly"
	"context_crawl/types"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitHubPipeline 用于处理GitHub仓库的爬虫管道
type GitHubPipeline struct {
	apiToken   string
	httpClient *http.Client
	baseURL    string
}

// NewGitHubPipeline 创建一个新的GitHubPipeline实例
func NewGitHubPipeline(apiToken string) *GitHubPipeline {
	return &GitHubPipeline{
		apiToken: apiToken,
		baseURL:  "https://api.github.com",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Match 匹配方法，检查URL是否是GitHub仓库
func (p *GitHubPipeline) Match(url string) bool {
	return strings.Contains(url, "github.com")
}

// Process 处理GitHub仓库爬取
func (p *GitHubPipeline) Process(input types.Type) (types.Type, error) {
	// 解析输入URL，提取搜索关键词
	query := p.extractQuery(input.Url)
	if query != "" {
		// 执行搜索
		results, err := p.searchGitHub(query)
		if err != nil {
			return types.Type{
				Url:  input.Url,
				Text: fmt.Sprintf("GitHub搜索失败: %v", err),
			}, nil
		}

		return types.Type{
			Url:  input.Url,
			Text: results,
		}, nil
	}

	// 检查是否是issue页面
	if p.isIssueURL(input.Url) {
		// 处理issue页面
		results, err := p.processIssuePage(input.Url)
		if err != nil {
			return types.Type{
				Url:  input.Url,
				Text: fmt.Sprintf("GitHub issue处理失败: %v", err),
			}, nil
		}

		return types.Type{
			Url:  input.Url,
			Text: results,
		}, nil
	}

	// 检查是否是其他GitHub讨论页面
	if p.isDiscussionURL(input.Url) {
		// 处理讨论页面
		results, err := p.processDiscussionPage(input.Url)
		if err != nil {
			return types.Type{
				Url:  input.Url,
				Text: fmt.Sprintf("GitHub讨论处理失败: %v", err),
			}, nil
		}

		return types.Type{
			Url:  input.Url,
			Text: results,
		}, nil
	}

	// 回退到普通爬取
	return p.fallbackToCollyCrawl(input)
}

// fallbackToCollyCrawl 回退到普通的colly爬取
func (p *GitHubPipeline) fallbackToCollyCrawl(input types.Type) (types.Type, error) {
	// 创建colly pipeline实例
	collyPipeline := colly.NewCollyPipeline()

	// 执行普通爬取
	result, err := collyPipeline.Process(input)
	if err != nil {
		// 如果普通爬取失败，返回默认信息
		return types.Type{
			Url:  input.Url,
			Text: fmt.Sprintf("GitHub仓库: %s", input.Url),
		}, nil
	}

	return result, nil
}

// isIssueURL 检查是否是GitHub issue页面
func (p *GitHubPipeline) isIssueURL(url string) bool {
	return strings.Contains(url, "github.com") && strings.Contains(url, "/issues/")
}

// isDiscussionURL 检查是否是GitHub讨论页面
func (p *GitHubPipeline) isDiscussionURL(url string) bool {
	return strings.Contains(url, "github.com") && strings.Contains(url, "/discussions/")
}

// processIssuePage 处理GitHub issue页面
func (p *GitHubPipeline) processIssuePage(url string) (string, error) {
	// 解析URL，提取owner、repo和issue编号
	owner, repo, issueNumber := p.parseIssueURL(url)
	if owner == "" || repo == "" || issueNumber == "" {
		return fmt.Sprintf("GitHub issue页面: %s", url), nil
	}

	// 调用GitHub API获取issue详情
	issue, err := p.getIssueDetails(owner, repo, issueNumber)
	if err != nil {
		return fmt.Sprintf("GitHub issue页面: %s\n获取详情失败: %v", url, err), nil
	}

	// 格式化issue详情
	return p.formatIssueDetails(issue), nil
}

// processDiscussionPage 处理GitHub讨论页面
func (p *GitHubPipeline) processDiscussionPage(url string) (string, error) {
	// 解析URL，提取owner、repo和discussion编号
	owner, repo, discussionNumber := p.parseDiscussionURL(url)
	if owner == "" || repo == "" || discussionNumber == "" {
		return fmt.Sprintf("GitHub讨论页面: %s", url), nil
	}

	// 调用GitHub API获取discussion详情
	discussion, err := p.getDiscussionDetails(owner, repo, discussionNumber)
	if err != nil {
		return fmt.Sprintf("GitHub讨论页面: %s\n获取详情失败: %v", url, err), nil
	}

	// 格式化discussion详情
	return p.formatDiscussionDetails(discussion), nil
}

// parseIssueURL 解析GitHub issue URL
func (p *GitHubPipeline) parseIssueURL(url string) (owner, repo, issueNumber string) {
	// 匹配格式: https://github.com/{owner}/{repo}/issues/{number}
	parts := strings.Split(url, "/")
	if len(parts) >= 7 && parts[3] != "" && parts[4] != "" && parts[6] != "" {
		return parts[3], parts[4], parts[6]
	}
	return "", "", ""
}

// parseDiscussionURL 解析GitHub discussion URL
func (p *GitHubPipeline) parseDiscussionURL(url string) (owner, repo, discussionNumber string) {
	// 匹配格式: https://github.com/{owner}/{repo}/discussions/{number}
	parts := strings.Split(url, "/")
	if len(parts) >= 7 && parts[3] != "" && parts[4] != "" && parts[6] != "" {
		return parts[3], parts[4], parts[6]
	}
	return "", "", ""
}

// getIssueDetails 获取GitHub issue详情
func (p *GitHubPipeline) getIssueDetails(owner, repo, issueNumber string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/issues/%s", owner, repo, issueNumber)
	body, err := p.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var issue map[string]interface{}
	if err := json.Unmarshal(body, &issue); err != nil {
		return nil, err
	}

	return issue, nil
}

// getDiscussionDetails 获取GitHub discussion详情
func (p *GitHubPipeline) getDiscussionDetails(owner, repo, discussionNumber string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("/repos/%s/%s/discussions/%s", owner, repo, discussionNumber)
	body, err := p.doRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	var discussion map[string]interface{}
	if err := json.Unmarshal(body, &discussion); err != nil {
		return nil, err
	}

	return discussion, nil
}

// formatIssueDetails 格式化GitHub issue详情
func (p *GitHubPipeline) formatIssueDetails(issue map[string]interface{}) string {
	title := "未知"
	if t, ok := issue["title"].(string); ok {
		title = t
	}

	body := ""
	if b, ok := issue["body"].(string); ok {
		body = b
	}

	state := "未知"
	if s, ok := issue["state"].(string); ok {
		state = s
	}

	user := "未知"
	if u, ok := issue["user"].(map[string]interface{}); ok {
		if login, ok := u["login"].(string); ok {
			user = login
		}
	}

	createdAt := "未知"
	if c, ok := issue["created_at"].(string); ok {
		createdAt = c
	}

	var result []string
	result = append(result, "## GitHub Issue详情")
	result = append(result, fmt.Sprintf("### 标题: %s", title))
	result = append(result, fmt.Sprintf("### 状态: %s", state))
	result = append(result, fmt.Sprintf("### 创建者: %s", user))
	result = append(result, fmt.Sprintf("### 创建时间: %s", createdAt))
	result = append(result, "### 内容:")
	result = append(result, body)

	return strings.Join(result, "\n\n")
}

// formatDiscussionDetails 格式化GitHub discussion详情
func (p *GitHubPipeline) formatDiscussionDetails(discussion map[string]interface{}) string {
	title := "未知"
	if t, ok := discussion["title"].(string); ok {
		title = t
	}

	body := ""
	if b, ok := discussion["body"].(string); ok {
		body = b
	}

	state := "未知"
	if s, ok := discussion["state"].(string); ok {
		state = s
	}

	user := "未知"
	if u, ok := discussion["user"].(map[string]interface{}); ok {
		if login, ok := u["login"].(string); ok {
			user = login
		}
	}

	createdAt := "未知"
	if c, ok := discussion["created_at"].(string); ok {
		createdAt = c
	}

	var result []string
	result = append(result, "## GitHub Discussion详情")
	result = append(result, fmt.Sprintf("### 标题: %s", title))
	result = append(result, fmt.Sprintf("### 状态: %s", state))
	result = append(result, fmt.Sprintf("### 创建者: %s", user))
	result = append(result, fmt.Sprintf("### 创建时间: %s", createdAt))
	result = append(result, "### 内容:")
	result = append(result, body)

	return strings.Join(result, "\n\n")
}

// extractQuery 从URL中提取搜索关键词
func (p *GitHubPipeline) extractQuery(inputUrl string) string {
	// 检查URL是否包含"search"字符串
	if strings.Contains(inputUrl, "search") {
		// 解析URL
		u, err := url.Parse(inputUrl)
		if err == nil {
			// 获取"q"参数的值
			q := u.Query().Get("q")
			if q != "" {
				// 替换+为空格，返回解码后的搜索关键词
				return strings.ReplaceAll(q, "+", " ")
			}
		}
	}

	// 检查URL是否是GitHub搜索URL的其他形式
	if strings.Contains(inputUrl, "github.com") {
		// 尝试从URL路径中提取搜索关键词
		// 例如：https://github.com/search?q=golang
		if strings.Contains(inputUrl, "?q=") {
			// 提取?q=后面的内容
			queryPart := strings.Split(inputUrl, "?q=")
			if len(queryPart) > 1 {
				// 移除可能的其他参数
				query := strings.Split(queryPart[1], "&")[0]
				if query != "" {
					// 替换+为空格，返回解码后的搜索关键词
					return strings.ReplaceAll(query, "+", " ")
				}
			}
		}
	}

	// 否则返回空字符串
	return ""
}

// searchGitHub 执行GitHub搜索
func (p *GitHubPipeline) searchGitHub(query string) (string, error) {
	var results []string

	// 搜索仓库
	repos, err := p.SearchRepositories(query, 1, 5)
	if err == nil {
		reposResult := p.formatRepositoriesResult(repos)
		if reposResult != "" {
			results = append(results, reposResult)
		}
	}

	// 搜索代码
	code, err := p.SearchCode(query, 1, 5)
	if err == nil {
		codeResult := p.formatCodeResult(code)
		if codeResult != "" {
			results = append(results, codeResult)
		}
	}

	// 搜索issues
	issues, err := p.SearchIssues(query, 1, 5)
	if err == nil {
		issuesResult := p.formatIssuesResult(issues)
		if issuesResult != "" {
			results = append(results, issuesResult)
		}
	}

	// 搜索文档（这里使用仓库搜索，过滤包含docs的仓库）
	docsQuery := fmt.Sprintf("%s docs", query)
	docs, err := p.SearchRepositories(docsQuery, 1, 5)
	if err == nil {
		docsResult := p.formatDocsResult(docs)
		if docsResult != "" {
			results = append(results, docsResult)
		}
	}

	if len(results) == 0 {
		return "没有找到相关结果", nil
	}

	return strings.Join(results, "\n\n"), nil
}

// formatRepositoriesResult 格式化仓库搜索结果
func (p *GitHubPipeline) formatRepositoriesResult(result map[string]interface{}) string {
	items, ok := result["items"].([]interface{})
	if !ok || len(items) == 0 {
		return ""
	}

	var repoResults []string
	repoResults = append(repoResults, "## GitHub仓库搜索结果")

	for _, item := range items {
		repo, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name := "未知"
		if n, ok := repo["name"].(string); ok {
			name = n
		}

		fullName := "未知"
		if fn, ok := repo["full_name"].(string); ok {
			fullName = fn
		}

		description := ""
		if d, ok := repo["description"].(string); ok {
			description = d
		}

		url := ""
		if u, ok := repo["html_url"].(string); ok {
			url = u
		}

		stars := 0
		if s, ok := repo["stargazers_count"].(float64); ok {
			stars = int(s)
		}

		repoResults = append(repoResults, fmt.Sprintf("- **%s** (%s)\n  描述: %s\n  链接: %s\n  星标: %d", name, fullName, description, url, stars))
	}

	return strings.Join(repoResults, "\n")
}

// formatCodeResult 格式化代码搜索结果
func (p *GitHubPipeline) formatCodeResult(result map[string]interface{}) string {
	items, ok := result["items"].([]interface{})
	if !ok || len(items) == 0 {
		return ""
	}

	var codeResults []string
	codeResults = append(codeResults, "## GitHub代码搜索结果")

	for _, item := range items {
		code, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		filename := "未知"
		if fn, ok := code["filename"].(string); ok {
			filename = fn
		}

		repoName := "未知"
		if repo, ok := code["repository"].(map[string]interface{}); ok {
			if rn, ok := repo["name"].(string); ok {
				repoName = rn
			}
		}

		path := ""
		if p, ok := code["path"].(string); ok {
			path = p
		}

		url := ""
		if u, ok := code["html_url"].(string); ok {
			url = u
		}

		codeResults = append(codeResults, fmt.Sprintf("- **%s** (在 %s 中)\n  路径: %s\n  链接: %s", filename, repoName, path, url))
	}

	return strings.Join(codeResults, "\n")
}

// formatIssuesResult 格式化issues搜索结果
func (p *GitHubPipeline) formatIssuesResult(result map[string]interface{}) string {
	items, ok := result["items"].([]interface{})
	if !ok || len(items) == 0 {
		return ""
	}

	var issuesResults []string
	issuesResults = append(issuesResults, "## GitHub Issues搜索结果")

	for _, item := range items {
		issue, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		title := "未知"
		if t, ok := issue["title"].(string); ok {
			title = t
		}

		repoName := "未知"
		if repo, ok := issue["repository"].(map[string]interface{}); ok {
			if rn, ok := repo["name"].(string); ok {
				repoName = rn
			}
		}

		url := ""
		if u, ok := issue["html_url"].(string); ok {
			url = u
		}

		state := "未知"
		if s, ok := issue["state"].(string); ok {
			state = s
		}

		issuesResults = append(issuesResults, fmt.Sprintf("- **%s** (在 %s 中)\n  状态: %s\n  链接: %s", title, repoName, state, url))
	}

	return strings.Join(issuesResults, "\n")
}

// formatDocsResult 格式化文档搜索结果
func (p *GitHubPipeline) formatDocsResult(result map[string]interface{}) string {
	items, ok := result["items"].([]interface{})
	if !ok || len(items) == 0 {
		return ""
	}

	var docsResults []string
	docsResults = append(docsResults, "## GitHub文档搜索结果")

	for _, item := range items {
		doc, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		name := "未知"
		if n, ok := doc["name"].(string); ok {
			name = n
		}

		fullName := "未知"
		if fn, ok := doc["full_name"].(string); ok {
			fullName = fn
		}

		description := ""
		if d, ok := doc["description"].(string); ok {
			description = d
		}

		url := ""
		if u, ok := doc["html_url"].(string); ok {
			url = u
		}

		docsResults = append(docsResults, fmt.Sprintf("- **%s** (%s)\n  描述: %s\n  链接: %s", name, fullName, description, url))
	}

	return strings.Join(docsResults, "\n")
}

// doRequest 执行HTTP请求
func (p *GitHubPipeline) doRequest(method, endpoint string, params map[string]string) ([]byte, error) {
	reqURL := fmt.Sprintf("%s%s", p.baseURL, endpoint)

	// 构建查询参数
	if len(params) > 0 {
		values := url.Values{}
		for k, v := range params {
			values.Add(k, v)
		}
		reqURL = fmt.Sprintf("%s?%s", reqURL, values.Encode())
	}

	// 创建请求
	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return nil, err
	}

	// 添加认证头
	if p.apiToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", p.apiToken))
	}

	// 添加User-Agent头
	req.Header.Set("User-Agent", "GitHub-Crawler")

	// 执行请求（带重试）
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err := p.httpClient.Do(req)
		if err != nil {
			// 网络错误，重试
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return nil, err
		}

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			// 读取错误，重试
			if i < maxRetries-1 {
				time.Sleep(time.Duration(i+1) * time.Second)
				continue
			}
			return nil, err
		}

		// 检查响应状态
		if resp.StatusCode == http.StatusOK {
			return body, nil
		}

		// 403 Forbidden（可能是限流），重试
		if resp.StatusCode == http.StatusForbidden && i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * 2 * time.Second)
			continue
		}

		// 其他错误，返回
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil, fmt.Errorf("max retries reached")
}

// SearchRepositories 搜索GitHub仓库
func (p *GitHubPipeline) SearchRepositories(query string, page, perPage int) (map[string]interface{}, error) {
	params := map[string]string{
		"q":        query,
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}

	body, err := p.doRequest("GET", "/search/repositories", params)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SearchCode 搜索GitHub代码
func (p *GitHubPipeline) SearchCode(query string, page, perPage int) (map[string]interface{}, error) {
	params := map[string]string{
		"q":        query,
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}

	body, err := p.doRequest("GET", "/search/code", params)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// SearchIssues 搜索GitHub issues
func (p *GitHubPipeline) SearchIssues(query string, page, perPage int) (map[string]interface{}, error) {
	params := map[string]string{
		"q":        query,
		"page":     fmt.Sprintf("%d", page),
		"per_page": fmt.Sprintf("%d", perPage),
	}

	body, err := p.doRequest("GET", "/search/issues", params)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
