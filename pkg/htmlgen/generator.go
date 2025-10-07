package htmlgen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/yourusername/sheduled-reports-app/pkg/grafana"
	"github.com/yourusername/sheduled-reports-app/pkg/model"
)

// Generator generates HTML reports from dashboard data
type Generator struct{}

// NewGenerator creates a new HTML generator
func NewGenerator() *Generator {
	return &Generator{}
}

// PanelData represents a panel with its query results
type PanelData struct {
	Panel   grafana.Panel
	Results []grafana.QueryResult
}

// ReportData holds all data needed to generate a report
type ReportData struct {
	Title       string
	Description string
	Generated   string
	TimeRange   string
	Panels      []PanelData
}

// Generate creates an HTML report from dashboard data
func (g *Generator) Generate(dashboard *grafana.Dashboard, panelsData []PanelData, schedule *model.Schedule) ([]byte, error) {
	reportData := ReportData{
		Title:       dashboard.Title,
		Description: fmt.Sprintf("Report generated from dashboard %s", dashboard.UID),
		Generated:   time.Now().Format(time.RFC1123),
		TimeRange:   fmt.Sprintf("%s to %s", schedule.RangeFrom, schedule.RangeTo),
		Panels:      panelsData,
	}

	tmpl, err := template.New("report").Funcs(template.FuncMap{
		"json": func(v interface{}) string {
			b, _ := json.Marshal(v)
			return string(b)
		},
		"chartType": g.getChartType,
		"chartData": g.extractChartData,
		"iterate": func(count int) []int {
			result := make([]int, count)
			for i := range result {
				result[i] = i
			}
			return result
		},
	}).Parse(htmlTemplate)

	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, reportData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

// getChartType maps Grafana panel types to Chart.js types
func (g *Generator) getChartType(panelType string) string {
	switch panelType {
	case "graph", "timeseries":
		return "line"
	case "bargauge", "barchart":
		return "bar"
	case "piechart":
		return "pie"
	case "stat", "gauge":
		return "doughnut"
	default:
		return "line"
	}
}

// extractChartData converts Grafana query results to Chart.js data format
func (g *Generator) extractChartData(results []grafana.QueryResult) map[string]interface{} {
	if len(results) == 0 {
		return map[string]interface{}{
			"labels":   []string{},
			"datasets": []interface{}{},
		}
	}

	var labels []interface{}
	datasets := []map[string]interface{}{}

	for _, result := range results {
		for _, frame := range result.Frames {
			// Find time and value fields
			var timeField, valueField *grafana.Field
			for i := range frame.Fields {
				field := &frame.Fields[i]
				if field.Type == "time" {
					timeField = field
				} else if field.Type == "number" {
					valueField = field
				}
			}

			if timeField == nil || valueField == nil {
				continue
			}

			// Extract labels (time values)
			if len(labels) == 0 {
				for _, v := range timeField.Values {
					if t, ok := v.(float64); ok {
						// Convert milliseconds to readable format
						timestamp := time.Unix(int64(t)/1000, 0)
						labels = append(labels, timestamp.Format("15:04:05"))
					} else if s, ok := v.(string); ok {
						labels = append(labels, s)
					}
				}
			}

			// Extract dataset
			dataset := map[string]interface{}{
				"label": frame.Name,
				"data":  valueField.Values,
				"borderColor": g.getColor(len(datasets)),
				"backgroundColor": g.getColorWithAlpha(len(datasets)),
				"borderWidth": 2,
				"fill": false,
			}
			datasets = append(datasets, dataset)
		}
	}

	return map[string]interface{}{
		"labels":   labels,
		"datasets": datasets,
	}
}

// getColor returns a color for a dataset index
func (g *Generator) getColor(index int) string {
	colors := []string{
		"rgb(54, 162, 235)",
		"rgb(255, 99, 132)",
		"rgb(75, 192, 192)",
		"rgb(255, 205, 86)",
		"rgb(153, 102, 255)",
		"rgb(255, 159, 64)",
	}
	return colors[index%len(colors)]
}

// getColorWithAlpha returns a color with alpha for a dataset index
func (g *Generator) getColorWithAlpha(index int) string {
	color := g.getColor(index)
	return strings.Replace(color, "rgb(", "rgba(", 1) + ", 0.2)"
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Report</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: white;
            color: #333;
            padding: 40px 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            border-bottom: 2px solid #f0f0f0;
            padding-bottom: 20px;
            margin-bottom: 40px;
        }
        h1 {
            font-size: 32px;
            margin-bottom: 10px;
            color: #000;
        }
        .meta {
            color: #666;
            font-size: 14px;
        }
        .meta-item {
            display: inline-block;
            margin-right: 20px;
        }
        .panel {
            margin-bottom: 40px;
            page-break-inside: avoid;
        }
        .panel-header {
            font-size: 20px;
            font-weight: 600;
            margin-bottom: 15px;
            color: #000;
        }
        .chart-container {
            position: relative;
            height: 400px;
            margin-bottom: 20px;
            border: 1px solid #e0e0e0;
            border-radius: 4px;
            padding: 20px;
            background: #fafafa;
        }
        .table-container {
            overflow-x: auto;
            border: 1px solid #e0e0e0;
            border-radius: 4px;
        }
        table {
            width: 100%;
            border-collapse: collapse;
            background: white;
        }
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #e0e0e0;
        }
        th {
            background: #f5f5f5;
            font-weight: 600;
        }
        tr:last-child td {
            border-bottom: none;
        }
        .no-data {
            padding: 40px;
            text-align: center;
            color: #999;
            font-style: italic;
        }
        @media print {
            body {
                padding: 0;
            }
            .panel {
                page-break-inside: avoid;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>{{.Title}}</h1>
            <div class="meta">
                <span class="meta-item"><strong>Generated:</strong> {{.Generated}}</span>
                <span class="meta-item"><strong>Time Range:</strong> {{.TimeRange}}</span>
            </div>
        </div>

        {{range $index, $panelData := .Panels}}
        <div class="panel">
            <div class="panel-header">{{$panelData.Panel.Title}}</div>

            {{if eq $panelData.Panel.Type "table"}}
                <div class="table-container">
                    {{if $panelData.Results}}
                        <table>
                            <thead>
                                <tr>
                                    {{range (index $panelData.Results 0).Frames}}
                                        {{range .Fields}}
                                            <th>{{.Name}}</th>
                                        {{end}}
                                    {{end}}
                                </tr>
                            </thead>
                            <tbody>
                                {{range (index $panelData.Results 0).Frames}}
                                    {{$frame := .}}
                                    {{if gt (len .Fields) 0}}
                                        {{$rowCount := len (index .Fields 0).Values}}
                                        {{range $rowIndex := iterate $rowCount}}
                                            <tr>
                                                {{range $frame.Fields}}
                                                    <td>{{index .Values $rowIndex}}</td>
                                                {{end}}
                                            </tr>
                                        {{end}}
                                    {{end}}
                                {{end}}
                            </tbody>
                        </table>
                    {{else}}
                        <div class="no-data">No data available</div>
                    {{end}}
                </div>
            {{else}}
                <div class="chart-container">
                    <canvas id="chart-{{$index}}"></canvas>
                </div>
                <script>
                    (function() {
                        const ctx = document.getElementById('chart-{{$index}}');
                        const chartData = {{chartData $panelData.Results | json}};
                        const chartType = '{{chartType $panelData.Panel.Type}}';

                        new Chart(ctx, {
                            type: chartType,
                            data: chartData,
                            options: {
                                responsive: true,
                                maintainAspectRatio: false,
                                plugins: {
                                    legend: {
                                        display: true,
                                        position: 'top',
                                    },
                                    title: {
                                        display: false
                                    }
                                },
                                scales: chartType === 'line' || chartType === 'bar' ? {
                                    y: {
                                        beginAtZero: true
                                    }
                                } : {}
                            }
                        });
                    })();
                </script>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>`
