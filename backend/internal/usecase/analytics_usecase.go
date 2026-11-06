package usecase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/gorm"
)

type AnalyticsUseCase interface {
	Query(question string) (*AnalyticsResponse, error)
}

type analyticsUseCase struct {
	db     *gorm.DB
	apiKey string
	model  string
}

func NewAnalyticsUseCase(db *gorm.DB, apiKey string, model string) AnalyticsUseCase {
	return &analyticsUseCase{
		db:     db,
		apiKey: apiKey,
		model:  model,
	}
}

type AnalyticsResponse struct {
	SQL       string                   `json:"sql"`
	ChartType string                   `json:"chart_type"`
	Summary   string                   `json:"summary"`
	Columns   []string                 `json:"columns"`
	Data      []map[string]interface{} `json:"data"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

var promptTemplate = `
You are an expert PostgreSQL developer for an HRIS and Sales application.
Your job is to translate natural language questions into valid PostgreSQL SELECT queries and suggest a visual representation.

DATABASE SCHEMA:
- public.users (id UUID, email VARCHAR, role VARCHAR, company_id UUID) -- role values: 'superadmin', 'hr_admin', 'employee', 'manager', 'admin_nota', 'supervisor_sales', 'cashier', 'admin_gudang'
- public.employees (id UUID, user_id UUID, company_id UUID, branch_id UUID, department_id UUID, job_position_id UUID, employee_id_number VARCHAR, first_name VARCHAR, last_name VARCHAR, gender VARCHAR, join_date DATE, employment_status VARCHAR, salary NUMERIC)
- public.attendance_logs (id UUID, employee_id UUID, shift_id UUID, clock_in_time TIMESTAMP, clock_out_time TIMESTAMP, status VARCHAR) -- status values: 'ON_TIME', 'LATE', 'ABSENT'
- public.sales_visits (id UUID, employee_id UUID, store_id UUID, check_in_time TIMESTAMP, check_out_time TIMESTAMP, type VARCHAR, notes TEXT) -- type values: 'CHECKIN', 'CHECKOUT'
- public.sales_transactions (id UUID, store_id UUID, employee_id UUID, receipt_no VARCHAR, total_amount DECIMAL, discount_amount DECIMAL, transaction_date DATE, status VARCHAR, payment_status VARCHAR, payment_method VARCHAR) -- status: 'PENDING', 'VERIFIED'. payment_status: 'UNPAID', 'PARTIAL', 'PAID'
- public.stores (id UUID, company_id UUID, assigned_employee_id UUID, name VARCHAR, address TEXT, category VARCHAR, is_active BOOLEAN)
- public.branches (id UUID, name VARCHAR, address TEXT)
- public.leave_requests (id UUID, employee_id UUID, start_date DATE, end_date DATE, status VARCHAR, reason TEXT) -- status values: 'PENDING', 'APPROVED', 'REJECTED'
- public.reimbursements (id UUID, employee_id UUID, title VARCHAR, amount DECIMAL, status VARCHAR, description TEXT) -- status values: 'PENDING', 'APPROVED', 'REJECTED'

RULES:
1. ONLY generate SELECT queries. Do NOT write INSERT, UPDATE, DELETE, DROP, or ALTER queries.
2. Return a JSON object with the following structure:
{
  "sql": "the PostgreSQL SELECT query",
  "chart_type": "bar" | "line" | "pie" | "table",
  "summary": "Brief description of the query target"
}
3. The SQL query MUST be valid PostgreSQL. Use correct joins (e.g. join public.employees on public.attendance_logs.employee_id = public.employees.id).
4. Ensure string matching uses case-insensitive patterns (ILIKE) if looking for names or regions.
5. Limit results to a maximum of 100 rows to prevent memory overload. Do not put semicolons in the SQL query.
6. The JSON object should be raw JSON string. Do not wrap in markdown code blocks.

Question: %s
JSON Response:
`

func (u *analyticsUseCase) Query(question string) (*AnalyticsResponse, error) {
	if u.apiKey == "" {
		return nil, errors.New("Gemini API key is not configured")
	}

	// 1. Call Gemini to translate to SQL
	prompt := fmt.Sprintf(promptTemplate, question)
	modelName := u.model
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}
	modelName = strings.TrimPrefix(modelName, "models/")
	apiURL := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", modelName, u.apiKey)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: prompt},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to make request to Gemini: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gemini response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var gResp geminiResponse
	if err := json.Unmarshal(bodyBytes, &gResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Gemini response: %v", err)
	}

	if len(gResp.Candidates) == 0 || len(gResp.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New("empty response from Gemini")
	}

	rawJSON := gResp.Candidates[0].Content.Parts[0].Text
	rawJSON = strings.TrimPrefix(rawJSON, "```json")
	rawJSON = strings.TrimPrefix(rawJSON, "```")
	rawJSON = strings.TrimSuffix(rawJSON, "```")
	rawJSON = strings.TrimSpace(rawJSON)

	var result struct {
		SQL       string `json:"sql"`
		ChartType string `json:"chart_type"`
		Summary   string `json:"summary"`
	}

	if err := json.Unmarshal([]byte(rawJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini SQL generation result: %v (raw response: %s)", err, rawJSON)
	}

	// 2. Validate SQL for security (SELECT queries only)
	cleanSQL := strings.TrimSpace(strings.ToUpper(result.SQL))
	if !strings.HasPrefix(cleanSQL, "SELECT") {
		return nil, fmt.Errorf("security validation failed: SQL query does not start with SELECT")
	}

	forbiddenKeywords := []string{
		"INSERT", "UPDATE", "DELETE", "DROP", "ALTER", "CREATE", "TRUNCATE", "RENAME", "GRANT", "REVOKE",
	}
	for _, kw := range forbiddenKeywords {
		if strings.Contains(cleanSQL, kw) {
			return nil, fmt.Errorf("security validation failed: SQL query contains forbidden keyword '%s'", kw)
		}
	}

	// 3. Execute SQL Query
	dbRows, err := u.db.Raw(result.SQL).Rows()
	if err != nil {
		return nil, fmt.Errorf("database query execution failed: %v", err)
	}
	defer dbRows.Close()

	cols, err := dbRows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	var dataList []map[string]interface{}
	for dbRows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := dbRows.Scan(columnPointers...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columns[i]
			b, ok := val.([]byte)
			if ok {
				rowMap[colName] = string(b)
			} else {
				if t, ok := val.(time.Time); ok {
					rowMap[colName] = t.Format("2006-01-02 15:04:05")
				} else {
					rowMap[colName] = val
				}
			}
		}
		dataList = append(dataList, rowMap)
	}

	return &AnalyticsResponse{
		SQL:       result.SQL,
		ChartType: result.ChartType,
		Summary:   result.Summary,
		Columns:   cols,
		Data:      dataList,
	}, nil
}
