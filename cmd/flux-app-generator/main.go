package main

import (
	"embed"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/EffectiveSloth/flux-app-generator/internal/generator"
	"github.com/EffectiveSloth/flux-app-generator/internal/helm"
	"github.com/EffectiveSloth/flux-app-generator/internal/types"
)

//go:embed templates
var templatesFS embed.FS

// loadTemplate loads a template from the embedded filesystem.
func loadTemplate(name string) (string, error) {
	data, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %w", name, err)
	}
	return string(data), nil
}

// Styles.
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#51CF66")).
			Bold(true)

	tableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#874BFD")).
			Padding(1, 0)

	tableHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(lipgloss.Color("#874BFD")).
				Bold(true).
				Padding(0, 1)

	tableRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	tableSelectedRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Background(lipgloss.Color("#F0F0F0")).
				Bold(true).
				Padding(0, 1)
)

type step int

const (
	stepAppName step = iota
	stepNamespace
	stepHelmRepoName
	stepHelmRepoURL
	stepChartPicker
	stepChartVersion
	stepInterval
	stepValuesPrefill // NEW: ask user about pre-filling values.yaml
	stepDone
)

type chartInfo struct {
	Name        string
	Description string
}

type model struct {
	step               step
	inputs             []textinput.Model
	intervalOpts       []string
	intervalIdx        int
	chartList          []chartInfo
	chartIdx           int
	chartVersions      []helm.ChartVersion
	versionIdx         int
	versionPage        int // Add pagination for versions
	pageSize           int // Number of versions per page
	quitting           bool
	result             string
	config             *types.AppConfig
	versionFetcher     *helm.VersionFetcher
	valuesPrefillIdx   int      // 0 = use default, 1 = empty
	valuesPrefillOpts  []string // ["Use default values", "Empty values file"]
	chartTarballValues string   // extracted values.yaml content
}

func initialModel() model {
	inputs := make([]textinput.Model, 6)
	placeholders := []string{
		"Application name",
		"Namespace (default: default)",
		"Helm repository name",
		"Helm repository URL",
		"", // unused, chart picker replaces this
		"Chart version (default: latest)",
	}
	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].Focus()
		inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4"))
		inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	}

	return model{
		step:               stepAppName,
		inputs:             inputs,
		intervalOpts:       []string{"1m", "5m", "10m", "30m", "1h"},
		intervalIdx:        1,
		chartList:          []chartInfo{},
		chartIdx:           0,
		chartVersions:      []helm.ChartVersion{},
		versionIdx:         0,
		versionPage:        0,
		pageSize:           10, // Show 10 versions per page
		config:             &types.AppConfig{},
		versionFetcher:     helm.NewVersionFetcher(),
		valuesPrefillOpts:  []string{"Use default values", "Empty values file"},
		valuesPrefillIdx:   0,
		chartTarballValues: "",
	}
}

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.step {
			case stepAppName:
				m.config.AppName = m.inputs[0].Value()
				m.step = stepNamespace
			case stepNamespace:
				val := m.inputs[1].Value()
				if val == "" {
					val = "default"
				}
				m.config.Namespace = val
				m.step = stepHelmRepoName
			case stepHelmRepoName:
				m.config.HelmRepoName = m.inputs[2].Value()
				m.step = stepHelmRepoURL
			case stepHelmRepoURL:
				m.config.HelmRepoURL = m.inputs[3].Value()
				// Fetch chart list from repo
				charts, err := m.versionFetcher.ListCharts(m.config.HelmRepoURL)
				if err != nil {
					m.result = "Error fetching charts: " + err.Error()
					return m, tea.Quit
				}
				m.chartList = make([]chartInfo, len(charts))
				for i, c := range charts {
					m.chartList[i] = chartInfo{Name: c.Name, Description: c.Description}
				}
				m.chartIdx = 0
				m.step = stepChartPicker
			case stepChartPicker:
				m.config.ChartName = m.chartList[m.chartIdx].Name
				// Fetch chart versions for selected chart
				versions, err := m.versionFetcher.FetchChartVersions(m.config.HelmRepoURL, m.config.ChartName)
				if err != nil {
					m.result = "Error fetching versions: " + err.Error()
					return m, tea.Quit
				}
				m.chartVersions = versions
				m.versionIdx = 0
				m.versionPage = 0
				m.step = stepChartVersion
			case stepChartVersion:
				if len(m.chartVersions) > 0 {
					// Calculate the actual version index from page and local index
					actualIdx := m.versionPage*m.pageSize + m.versionIdx
					if actualIdx < len(m.chartVersions) {
						m.config.ChartVersion = m.chartVersions[actualIdx].ChartVersion
					} else {
						val := m.inputs[5].Value()
						if val == "" {
							val = "latest"
						}
						m.config.ChartVersion = val
					}
				} else {
					val := m.inputs[5].Value()
					if val == "" {
						val = "latest"
					}
					m.config.ChartVersion = val
				}
				m.step = stepInterval
			case stepInterval:
				m.config.Interval = m.intervalOpts[m.intervalIdx]
				m.step = stepValuesPrefill
			case stepValuesPrefill:
				// After user chooses, proceed to file generation
				// Download and extract values.yaml if needed
				if m.valuesPrefillIdx == 0 {
					// Download and extract values.yaml
					values, err := helm.DownloadAndExtractValuesYAML(m.config.HelmRepoURL, m.config.ChartName, m.config.ChartVersion)
					if err != nil {
						m.result = "Error downloading chart values.yaml: " + err.Error()
						return m, tea.Quit
					}
					m.chartTarballValues = values
				} else {
					m.chartTarballValues = ""
				}
				// Load templates from embedded filesystem
				var err error
				generator.HelmRepositoryTemplate, err = loadTemplate("helm-repository.yaml.tmpl")
				if err != nil {
					m.result = "Error loading template: " + err.Error()
					return m, tea.Quit
				}
				generator.HelmReleaseTemplate, err = loadTemplate("helm-release.yaml.tmpl")
				if err != nil {
					m.result = "Error loading template: " + err.Error()
					return m, tea.Quit
				}
				generator.KustomizationTemplate, err = loadTemplate("kustomization.yaml.tmpl")
				if err != nil {
					m.result = "Error loading template: " + err.Error()
					return m, tea.Quit
				}
				// Pass the extracted or empty values to the generator
				m.config.Values = map[string]interface{}{"__raw_yaml__": m.chartTarballValues}
				err = generator.GenerateFluxStructure(m.config)
				if err != nil {
					m.result = "Error: " + err.Error()
				} else {
					m.result = "✅ Files generated!"
				}
				return m, tea.Quit
			}
		case "up", "k":
			if m.step == stepInterval && m.intervalIdx > 0 {
				m.intervalIdx--
			} else if m.step == stepChartVersion && m.versionIdx > 0 {
				m.versionIdx--
			} else if m.step == stepChartPicker && m.chartIdx > 0 {
				m.chartIdx--
			} else if m.step == stepValuesPrefill && m.valuesPrefillIdx > 0 {
				m.valuesPrefillIdx--
			}
		case "down", "j":
			if m.step == stepInterval && m.intervalIdx < len(m.intervalOpts)-1 {
				m.intervalIdx++
			} else if m.step == stepChartVersion {
				// Check if we can go down within the current page
				startIdx := m.versionPage * m.pageSize
				endIdx := startIdx + m.pageSize
				if endIdx > len(m.chartVersions) {
					endIdx = len(m.chartVersions)
				}
				if m.versionIdx < endIdx-startIdx-1 {
					m.versionIdx++
				}
			} else if m.step == stepChartPicker && m.chartIdx < len(m.chartList)-1 {
				m.chartIdx++
			} else if m.step == stepValuesPrefill && m.valuesPrefillIdx < len(m.valuesPrefillOpts)-1 {
				m.valuesPrefillIdx++
			}
		case "left", "h":
			if m.step == stepChartVersion && m.versionPage > 0 {
				m.versionPage--
				m.versionIdx = 0 // Reset to first item on new page
			}
		case "right", "l":
			if m.step == stepChartVersion {
				nextPage := m.versionPage + 1
				if nextPage*m.pageSize < len(m.chartVersions) {
					m.versionPage = nextPage
					m.versionIdx = 0 // Reset to first item on new page
				}
			}
		}
	}

	// Handle text input updates
	switch m.step {
	case stepAppName:
		var cmd tea.Cmd
		m.inputs[0], cmd = m.inputs[0].Update(msg)
		return m, cmd
	case stepNamespace:
		var cmd tea.Cmd
		m.inputs[1], cmd = m.inputs[1].Update(msg)
		return m, cmd
	case stepHelmRepoName:
		var cmd tea.Cmd
		m.inputs[2], cmd = m.inputs[2].Update(msg)
		return m, cmd
	case stepHelmRepoURL:
		var cmd tea.Cmd
		m.inputs[3], cmd = m.inputs[3].Update(msg)
		return m, cmd
	case stepChartVersion:
		var cmd tea.Cmd
		m.inputs[5], cmd = m.inputs[5].Update(msg)
		return m, cmd
	}

	return m, nil
}

func renderTable(headers []string, rows [][]string, selectedIdx int) string {
	if len(rows) == 0 {
		return "No items available"
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Build table
	var table string

	// Header
	headerRow := ""
	for i, header := range headers {
		headerRow += tableHeaderStyle.Render(fmt.Sprintf("%-*s", colWidths[i], header))
		if i < len(headers)-1 {
			headerRow += " "
		}
	}
	table += headerRow + "\n"

	// Rows
	for i, row := range rows {
		rowStr := ""
		for j, cell := range row {
			style := tableRowStyle
			if i == selectedIdx {
				style = tableSelectedRowStyle
			}
			rowStr += style.Render(fmt.Sprintf("%-*s", colWidths[j], cell))
			if j < len(row)-1 {
				rowStr += " "
			}
		}
		table += rowStr + "\n"
	}

	return tableStyle.Render(table)
}

func (m *model) View() string {
	if m.quitting {
		return titleStyle.Render("👋 Goodbye!") + "\n"
	}
	if m.result != "" {
		if m.result == "✅ Files generated!" {
			return successStyle.Render(m.result) + "\n"
		}
		return errorStyle.Render(m.result) + "\n"
	}

	var content string

	switch m.step {
	case stepAppName:
		content = titleStyle.Render("🚀 Flux App Generator") + "\n\n"
		content += subtitleStyle.Render("Let's create your Flux GitOps structure") + "\n\n"
		content += "App Name: " + m.inputs[0].View() + "\n"
		content += subtitleStyle.Render("(enter to continue)")
	case stepNamespace:
		content = titleStyle.Render("📁 Namespace") + "\n\n"
		content += "Namespace: " + m.inputs[1].View() + "\n"
		content += subtitleStyle.Render("(enter to continue)")
	case stepHelmRepoName:
		content = titleStyle.Render("📦 Helm Repository") + "\n\n"
		content += "Helm Repo Name: " + m.inputs[2].View() + "\n"
		content += subtitleStyle.Render("(enter to continue)")
	case stepHelmRepoURL:
		content = titleStyle.Render("🔗 Repository URL") + "\n\n"
		content += "Helm Repo URL: " + m.inputs[3].View() + "\n"
		content += subtitleStyle.Render("(enter to continue)")
	case stepChartPicker:
		content = titleStyle.Render("📋 Available Charts") + "\n\n"
		if len(m.chartList) > 0 {
			headers := []string{"Chart Name", "Description"}
			rows := make([][]string, len(m.chartList))
			for i, chart := range m.chartList {
				rows[i] = []string{chart.Name, chart.Description}
			}
			content += renderTable(headers, rows, m.chartIdx)
		} else {
			content += "No charts found in repository"
		}
		content += "\n" + subtitleStyle.Render("(up/down to navigate, enter to select)")
	case stepChartVersion:
		content = titleStyle.Render("🏷️  Chart Version") + "\n\n"
		if len(m.chartVersions) > 0 {
			// Calculate pagination
			startIdx := m.versionPage * m.pageSize
			endIdx := startIdx + m.pageSize
			if endIdx > len(m.chartVersions) {
				endIdx = len(m.chartVersions)
			}

			// Show pagination info
			totalPages := (len(m.chartVersions) + m.pageSize - 1) / m.pageSize
			content += subtitleStyle.Render(fmt.Sprintf("Page %d of %d (%d total versions)", m.versionPage+1, totalPages, len(m.chartVersions))) + "\n\n"

			// Show current page of versions
			pageVersions := m.chartVersions[startIdx:endIdx]
			headers := []string{"Chart Version", "App Version", "Description"}
			rows := make([][]string, len(pageVersions))
			for i, version := range pageVersions {
				rows[i] = []string{version.ChartVersion, version.AppVersion, version.Description}
			}
			content += renderTable(headers, rows, m.versionIdx)

			// Show navigation hints
			navHints := subtitleStyle.Render("(up/down to navigate, left/right to change page, enter to select)")
			if totalPages > 1 {
				navHints += "\n" + subtitleStyle.Render(fmt.Sprintf("Showing versions %d-%d of %d", startIdx+1, endIdx, len(m.chartVersions)))
			}
			content += "\n" + navHints
		} else {
			content += "Chart Version: " + m.inputs[5].View() + "\n"
			content += "\n" + subtitleStyle.Render("(enter to continue)")
		}
	case stepInterval:
		content = titleStyle.Render("⏱️  Sync Interval") + "\n\n"
		content += "Interval: "
		for i, opt := range m.intervalOpts {
			if i == m.intervalIdx {
				content += selectedStyle.Render("> " + opt + " <")
			} else {
				content += "  " + opt + "  "
			}
		}
		content += "\n" + subtitleStyle.Render("(up/down to change, enter to continue)")
	case stepValuesPrefill:
		content = titleStyle.Render("📝 Pre-fill Values File?") + "\n\n"
		content += "Do you want to pre-fill helm-values.yaml with the chart's default values.yaml, or generate an empty one?\n\n"
		for i, opt := range m.valuesPrefillOpts {
			if i == m.valuesPrefillIdx {
				content += selectedStyle.Render("> "+opt+" <") + "\n"
			} else {
				content += "  " + opt + "  \n"
			}
		}
		content += subtitleStyle.Render("(up/down to choose, enter to continue)")
	}
	return content
}

func main() {
	m := initialModel()
	if _, err := tea.NewProgram(&m).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
