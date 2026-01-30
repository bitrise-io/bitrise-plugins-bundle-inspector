package report

// htmlTemplate is the embedded HTML template for the interactive report
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        :root {
            /* Light theme colors */
            --bg-primary: #f5f5f7;
            --bg-secondary: #ffffff;
            --text-primary: #1d1d1f;
            --text-secondary: #6e6e73;
            --text-tertiary: #86868b;
            --border-color: #e5e5e7;
            --shadow: rgba(0, 0, 0, 0.1);
            --shadow-hover: rgba(0, 0, 0, 0.15);
        }

        [data-theme="dark"] {
            /* Dark theme colors */
            --bg-primary: #1a1a1a;
            --bg-secondary: #2a2a2a;
            --text-primary: #f5f5f7;
            --text-secondary: #a1a1a6;
            --text-tertiary: #86868b;
            --border-color: #3a3a3a;
            --shadow: rgba(0, 0, 0, 0.3);
            --shadow-hover: rgba(0, 0, 0, 0.4);
        }

        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.6;
            transition: background-color 0.3s ease, color 0.3s ease;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }

        header {
            background: var(--bg-secondary);
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px var(--shadow);
            transition: background-color 0.3s ease, box-shadow 0.3s ease;
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            flex-wrap: wrap;
            gap: 20px;
        }

        .header-content {
            flex: 1;
        }

        h1 {
            font-size: 32px;
            font-weight: 600;
            margin-bottom: 10px;
            color: var(--text-primary);
        }

        .artifact-info {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            margin-top: 15px;
            color: var(--text-secondary);
        }

        .artifact-info span {
            display: flex;
            align-items: center;
            gap: 5px;
        }

        .theme-toggle {
            background: var(--bg-primary);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            padding: 10px 16px;
            cursor: pointer;
            font-size: 14px;
            color: var(--text-primary);
            display: flex;
            align-items: center;
            gap: 8px;
            transition: all 0.3s ease;
        }

        .theme-toggle:hover {
            background: var(--border-color);
            transform: translateY(-1px);
            box-shadow: 0 2px 8px var(--shadow-hover);
        }

        .theme-icon {
            font-size: 18px;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }

        .metric-card {
            background: var(--bg-secondary);
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px var(--shadow);
            transition: background-color 0.3s ease, box-shadow 0.3s ease;
        }

        .metric-label {
            font-size: 14px;
            color: var(--text-secondary);
            margin-bottom: 8px;
        }

        .metric-value {
            font-size: 28px;
            font-weight: 600;
            color: var(--text-primary);
        }

        .metric-secondary {
            font-size: 14px;
            color: var(--text-tertiary);
            margin-top: 4px;
        }

        .main-content {
            display: grid;
            grid-template-columns: 1fr 350px;
            gap: 20px;
            margin-bottom: 20px;
        }

        @media (max-width: 1024px) {
            .main-content {
                grid-template-columns: 1fr;
            }
        }

        .treemap-container {
            background: var(--bg-secondary);
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px var(--shadow);
            min-height: 600px;
            transition: background-color 0.3s ease, box-shadow 0.3s ease;
        }

        .treemap-container h2 {
            color: var(--text-primary);
            margin-bottom: 15px;
        }

        #treemap {
            width: 100%;
            height: 600px;
        }

        .side-panel {
            display: flex;
            flex-direction: column;
            gap: 20px;
        }

        .chart-card {
            background: var(--bg-secondary);
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px var(--shadow);
            transition: background-color 0.3s ease, box-shadow 0.3s ease;
        }

        .chart-card h2 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 15px;
            color: var(--text-primary);
        }

        .chart {
            width: 100%;
            height: 280px;
        }

        .insights-section {
            background: var(--bg-secondary);
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 8px var(--shadow);
            transition: background-color 0.3s ease, box-shadow 0.3s ease;
        }

        .insights-section h2 {
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 20px;
            color: var(--text-primary);
        }

        .insight-card {
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 0;
            margin-bottom: 20px;
            background: var(--bg-primary);
            transition: all 0.3s ease;
            overflow: hidden;
        }

        .insight-card:last-child {
            margin-bottom: 0;
        }

        .insight-card:hover {
            box-shadow: 0 4px 12px var(--shadow-hover);
            transform: translateY(-2px);
        }

        .insight-header {
            padding: 20px;
            display: flex;
            align-items: flex-start;
            gap: 15px;
            cursor: pointer;
            user-select: none;
        }

        .insight-icon {
            font-size: 32px;
            line-height: 1;
            flex-shrink: 0;
        }

        .insight-content {
            flex: 1;
            min-width: 0;
        }

        .insight-title {
            font-size: 18px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 8px;
        }

        .insight-description {
            font-size: 14px;
            color: var(--text-secondary);
            margin-bottom: 12px;
            line-height: 1.5;
        }

        .insight-meta {
            display: flex;
            align-items: center;
            gap: 15px;
            flex-wrap: wrap;
        }

        .insight-savings {
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .savings-amount {
            font-size: 16px;
            font-weight: 700;
            color: #34c759;
        }

        .savings-percentage {
            font-size: 14px;
            font-weight: 600;
            color: #34c759;
            background: rgba(52, 199, 89, 0.1);
            padding: 4px 8px;
            border-radius: 4px;
        }

        .insight-count {
            font-size: 13px;
            color: var(--text-secondary);
            padding: 4px 10px;
            background: var(--bg-secondary);
            border-radius: 12px;
        }

        .learn-more-link {
            font-size: 13px;
            color: #007aff;
            text-decoration: none;
            font-weight: 500;
            display: inline-flex;
            align-items: center;
            gap: 4px;
        }

        .learn-more-link:hover {
            text-decoration: underline;
        }

        .expand-indicator {
            font-size: 20px;
            color: var(--text-secondary);
            transition: transform 0.3s ease;
            flex-shrink: 0;
        }

        .insight-card.expanded .expand-indicator {
            transform: rotate(180deg);
        }

        .insight-files {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease;
            border-top: 1px solid var(--border-color);
        }

        .insight-card.expanded .insight-files {
            max-height: 400px;
            overflow-y: auto;
        }

        .insight-files-content {
            padding: 20px;
        }

        .insight-files-header {
            font-size: 14px;
            font-weight: 600;
            color: var(--text-primary);
            margin-bottom: 12px;
        }

        .files-list {
            list-style: none;
            margin: 0;
            padding: 0;
        }

        .files-list li {
            padding: 8px 12px;
            color: var(--text-secondary);
            font-family: 'Monaco', 'Courier New', monospace;
            font-size: 12px;
            background: var(--bg-secondary);
            border-radius: 6px;
            margin-bottom: 6px;
            transition: background-color 0.2s ease;
        }

        .files-list li:hover {
            background: var(--border-color);
        }

        .files-list li:last-child {
            margin-bottom: 0;
        }

        .no-insights {
            text-align: center;
            padding: 40px;
            color: #34c759;
            font-size: 18px;
        }

        .search-container {
            margin-bottom: 15px;
        }

        #search-input {
            width: 100%%;
            padding: 10px 15px;
            border: 1px solid var(--border-color);
            border-radius: 8px;
            font-size: 14px;
            outline: none;
            transition: border-color 0.2s, background-color 0.3s ease, color 0.3s ease;
            background: var(--bg-primary);
            color: var(--text-primary);
        }

        #search-input:focus {
            border-color: #007aff;
        }

        #search-input::placeholder {
            color: var(--text-tertiary);
        }

        .legend {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid var(--border-color);
        }

        .legend-item {
            display: flex;
            align-items: center;
            gap: 5px;
            font-size: 12px;
            color: var(--text-secondary);
        }

        .legend-color {
            width: 16px;
            height: 16px;
            border-radius: 3px;
        }

        footer {
            text-align: center;
            padding: 20px;
            color: var(--text-tertiary);
            font-size: 14px;
        }

        footer a {
            color: #007aff;
            text-decoration: none;
        }

        footer a:hover {
            text-decoration: underline;
        }

        .warning-banner {
            background: #fff3cd;
            border: 1px solid #ffeaa7;
            border-radius: 8px;
            padding: 15px;
            margin-bottom: 20px;
            color: #856404;
        }

        .warning-banner strong {
            display: block;
            margin-bottom: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="header-content">
                <h1>{{.Title}}</h1>
                <div class="artifact-info">
                    <span><strong>Artifact:</strong> {{.ArtifactName}}</span>
                    <span><strong>Type:</strong> {{.ArtifactType}}</span>
                    <span><strong>Analyzed:</strong> <time>{{.Timestamp}}</time></span>
                </div>
            </div>
            <button class="theme-toggle" onclick="toggleTheme()" aria-label="Toggle theme">
                <span class="theme-icon">🌙</span>
                <span class="theme-text">Dark Mode</span>
            </button>
        </header>

        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-label">Install Size</div>
                <div class="metric-value">{{.UncompressedSize}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Download Size</div>
                <div class="metric-value">{{.TotalSize}}</div>
                <div class="metric-secondary">Compression: {{.CompressionRatio}}</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Potential Savings</div>
                <div class="metric-value">{{.TotalSavings}}</div>
                <div class="metric-secondary">{{.SavingsPercentage}} of total</div>
            </div>
            <div class="metric-card">
                <div class="metric-label">Total Files</div>
                <div class="metric-value">{{.NodeCount}}</div>
            </div>
        </div>

        <div class="main-content">
            <div class="treemap-container">
                <h2>Bundle Treemap</h2>
                <div class="search-container">
                    <input type="text" id="search-input" placeholder="Search files (e.g., .png, Framework, Assets.car)">
                </div>
                <div id="treemap"></div>
                <div class="legend">
                    <div class="legend-item">
                        <div class="legend-color" style="background: #5470c6;"></div>
                        <span>Frameworks</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #91cc75;"></div>
                        <span>Libraries</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #fac858;"></div>
                        <span>Images</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #ea7ccc;"></div>
                        <span>Native Libs</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #73c0de;"></div>
                        <span>Asset Catalogs</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #3ba272;"></div>
                        <span>Resources</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #fc8452;"></div>
                        <span>UI</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #9a60b4;"></div>
                        <span>DEX</span>
                    </div>
                    <div class="legend-item">
                        <div class="legend-color" style="background: #999999;"></div>
                        <span>Other</span>
                    </div>
                </div>
            </div>

            <div class="side-panel">
                <div class="chart-card">
                    <h2>Category Breakdown</h2>
                    <div id="category-chart" class="chart"></div>
                </div>
                <div class="chart-card">
                    <h2>Top Extensions</h2>
                    <div id="extension-chart" class="chart"></div>
                </div>
            </div>
        </div>

        <div class="insights-section" id="insights-section">
            <h2>💡 Insights & Optimization Opportunities</h2>
            <div id="insights-list"></div>
        </div>

        <footer>
            Generated by Bundle Inspector | <a href="https://github.com/bitrise-io/bitrise-plugins-bundle-inspector" target="_blank">GitHub</a>
        </footer>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js"></script>
    <script>
        const reportData = {{.DataJSON}};

        // Theme management
        let currentTheme = 'light';
        let treemapChart = null;
        let categoryChart = null;
        let extensionChart = null;

        // Initialize theme from localStorage
        function initTheme() {
            const savedTheme = localStorage.getItem('theme') || 'light';
            currentTheme = savedTheme;
            document.documentElement.setAttribute('data-theme', savedTheme);
            updateThemeButton();
        }

        // Toggle theme
        function toggleTheme() {
            currentTheme = currentTheme === 'light' ? 'dark' : 'light';
            document.documentElement.setAttribute('data-theme', currentTheme);
            localStorage.setItem('theme', currentTheme);
            updateThemeButton();
            updateChartsTheme();
        }

        // Update theme button text and icon
        function updateThemeButton() {
            const button = document.querySelector('.theme-toggle');
            const icon = button.querySelector('.theme-icon');
            const text = button.querySelector('.theme-text');

            if (currentTheme === 'dark') {
                icon.textContent = '☀️';
                text.textContent = 'Light Mode';
            } else {
                icon.textContent = '🌙';
                text.textContent = 'Dark Mode';
            }
        }

        // Update all charts with new theme
        function updateChartsTheme() {
            if (treemapChart) {
                const treemapOption = getTreemapOption(reportData.fileTree);
                treemapChart.setOption(treemapOption, true);
            }
            if (categoryChart) {
                const categoryOption = getCategoryChartOption(reportData.categories);
                categoryChart.setOption(categoryOption, true);
            }
            if (extensionChart) {
                const extensionOption = getExtensionChartOption(reportData.extensions);
                extensionChart.setOption(extensionOption, true);
            }
        }

        // Get theme colors for ECharts
        function getThemeColors() {
            const isDark = currentTheme === 'dark';
            return {
                textColor: isDark ? '#f5f5f7' : '#1d1d1f',
                backgroundColor: isDark ? 'transparent' : 'transparent',
                axisLineColor: isDark ? '#3a3a3a' : '#e5e5e7',
                splitLineColor: isDark ? '#3a3a3a' : '#e5e5e7',
            };
        }

        // Format bytes helper
        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return (bytes / Math.pow(k, i)).toFixed(1) + ' ' + sizes[i];
        }

        // Color mapping for file types
        const fileTypeColors = {
            'framework': '#5470c6',
            'library': '#91cc75',
            'native': '#ea7ccc',
            'image': '#fac858',
            'asset_catalog': '#73c0de',
            'resource': '#3ba272',
            'ui': '#fc8452',
            'dex': '#9a60b4',
            'font': '#ee6666',
            'other': '#999999'
        };

        // Get color for file type
        function getColorForFileType(fileType) {
            return fileTypeColors[fileType] || fileTypeColors['other'];
        }

        // Apply colors to tree nodes
        function applyColorsToTree(node) {
            if (node.fileType) {
                node.itemStyle = {
                    color: getColorForFileType(node.fileType)
                };
            }
            if (node.children) {
                node.children.forEach(child => applyColorsToTree(child));
            }
        }

        // Get treemap option with theme support
        function getTreemapOption(data) {
            // Apply colors
            applyColorsToTree(data);

            const themeColors = getThemeColors();
            const isDark = currentTheme === 'dark';
            const borderColor = isDark ? '#666' : '#fff';
            const emphasisBorder = isDark ? '#fff' : '#333';
            const breadcrumbText = isDark ? '#f5f5f7' : '#333';

            return {
                tooltip: {
                    backgroundColor: isDark ? '#2a2a2a' : '#fff',
                    borderColor: isDark ? '#3a3a3a' : '#e5e5e7',
                    textStyle: {
                        color: themeColors.textColor
                    },
                    formatter: function(info) {
                        const value = info.value;
                        const name = info.name;
                        const path = info.data.path || name;
                        const treePathInfo = info.treePathInfo || [];
                        let percentage = '0.0';

                        if (treePathInfo.length > 0) {
                            const rootValue = treePathInfo[0].value;
                            if (rootValue > 0) {
                                percentage = ((value / rootValue) * 100).toFixed(2);
                            }
                        }

                        return '<strong>' + name + '</strong><br/>' +
                               'Path: ' + path + '<br/>' +
                               'Size: ' + formatBytes(value) + '<br/>' +
                               percentage + '%% of total';
                    }
                },
                series: [{
                    type: 'treemap',
                    width: '100%%',
                    height: '100%%',
                    roam: false,
                    nodeClick: 'zoomToNode',
                    breadcrumb: {
                        show: true,
                        height: 25,
                        itemStyle: {
                            color: isDark ? 'rgba(42,42,42,0.9)' : 'rgba(255,255,255,0.7)',
                            borderColor: isDark ? 'rgba(58,58,58,0.9)' : 'rgba(255,255,255,0.7)',
                            textStyle: {
                                color: breadcrumbText
                            }
                        },
                        emphasis: {
                            itemStyle: {
                                color: isDark ? 'rgba(42,42,42,1)' : 'rgba(255,255,255,1)',
                                textStyle: {
                                    color: breadcrumbText
                                }
                            }
                        }
                    },
                    label: {
                        show: true,
                        formatter: '{b}',
                        fontSize: 11,
                        overflow: 'truncate',
                        color: isDark ? '#fff' : '#000'
                    },
                    upperLabel: {
                        show: true,
                        height: 25,
                        color: '#fff',
                        textBorderColor: 'transparent'
                    },
                    itemStyle: {
                        borderColor: borderColor,
                        borderWidth: 2,
                        gapWidth: 2
                    },
                    emphasis: {
                        label: {
                            show: true,
                            fontSize: 12,
                            fontWeight: 'bold'
                        },
                        itemStyle: {
                            borderColor: emphasisBorder,
                            borderWidth: 3
                        }
                    },
                    levels: [
                        {
                            itemStyle: {
                                borderWidth: 0,
                                gapWidth: 5
                            }
                        },
                        {
                            itemStyle: {
                                gapWidth: 1
                            }
                        },
                        {
                            itemStyle: {
                                gapWidth: 1
                            }
                        }
                    ],
                    data: [data]
                }]
            };
        }

        // Create treemap visualization
        function createTreemap(data) {
            const container = document.getElementById('treemap');
            const chart = echarts.init(container, null, {
                renderer: 'canvas',
                useDirtyRect: true
            });

            const option = getTreemapOption(data);
            chart.setOption(option);

            window.addEventListener('resize', function() {
                chart.resize();
            });

            return chart;
        }

        // Get category chart option with theme support
        function getCategoryChartOption(categories) {
            const themeColors = getThemeColors();
            const isDark = currentTheme === 'dark';

            return {
                tooltip: {
                    trigger: 'item',
                    formatter: '{b}: {c} ({d}%%)',
                    backgroundColor: isDark ? '#2a2a2a' : '#fff',
                    borderColor: isDark ? '#3a3a3a' : '#e5e5e7',
                    textStyle: {
                        color: themeColors.textColor
                    }
                },
                series: [{
                    type: 'pie',
                    radius: ['40%%', '70%%'],
                    avoidLabelOverlap: true,
                    itemStyle: {
                        borderRadius: 8,
                        borderColor: isDark ? '#2a2a2a' : '#fff',
                        borderWidth: 2
                    },
                    label: {
                        show: true,
                        formatter: function(params) {
                            return params.name + '\n' + formatBytes(params.value);
                        },
                        fontSize: 11,
                        color: themeColors.textColor
                    },
                    emphasis: {
                        label: {
                            show: true,
                            fontSize: 13,
                            fontWeight: 'bold'
                        }
                    },
                    data: categories.map(cat => ({
                        name: cat.name,
                        value: cat.value
                    }))
                }]
            };
        }

        // Create category donut chart
        function createCategoryChart(categories) {
            const container = document.getElementById('category-chart');
            const chart = echarts.init(container);

            const option = getCategoryChartOption(categories);
            chart.setOption(option);

            window.addEventListener('resize', () => chart.resize());
            return chart;
        }

        // Get extension chart option with theme support
        function getExtensionChartOption(extensions) {
            const themeColors = getThemeColors();
            const isDark = currentTheme === 'dark';

            return {
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' },
                    formatter: function(params) {
                        const data = params[0];
                        return data.name + ': ' + formatBytes(data.value);
                    },
                    backgroundColor: isDark ? '#2a2a2a' : '#fff',
                    borderColor: isDark ? '#3a3a3a' : '#e5e5e7',
                    textStyle: {
                        color: themeColors.textColor
                    }
                },
                grid: {
                    left: '5%%',
                    right: '5%%',
                    bottom: '3%%',
                    top: '3%%',
                    containLabel: true
                },
                xAxis: {
                    type: 'value',
                    axisLabel: {
                        formatter: (value) => formatBytes(value),
                        fontSize: 10,
                        color: themeColors.textColor
                    },
                    axisLine: {
                        lineStyle: {
                            color: themeColors.axisLineColor
                        }
                    },
                    splitLine: {
                        lineStyle: {
                            color: themeColors.splitLineColor
                        }
                    }
                },
                yAxis: {
                    type: 'category',
                    data: extensions.map(e => e.name),
                    axisLabel: {
                        fontSize: 10,
                        color: themeColors.textColor
                    },
                    axisLine: {
                        lineStyle: {
                            color: themeColors.axisLineColor
                        }
                    }
                },
                series: [{
                    type: 'bar',
                    data: extensions.map(e => e.value),
                    itemStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 1, 0, [
                            { offset: 0, color: '#5470c6' },
                            { offset: 1, color: '#91cc75' }
                        ])
                    },
                    label: {
                        show: true,
                        position: 'right',
                        formatter: (params) => formatBytes(params.value),
                        fontSize: 10,
                        color: themeColors.textColor
                    }
                }]
            };
        }

        // Create extension bar chart
        function createExtensionChart(extensions) {
            const container = document.getElementById('extension-chart');
            const chart = echarts.init(container);

            const option = getExtensionChartOption(extensions);
            chart.setOption(option);

            window.addEventListener('resize', () => chart.resize());
            return chart;
        }

        // Category metadata with icons and learn more links
        const categoryMetadata = {
            'strip-symbols': {
                icon: '🔧',
                title: 'Strip Binary Symbols',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/strip-debug-symbols.html'
            },
            'frameworks': {
                icon: '📦',
                title: 'Unused Frameworks',
                learnMore: 'https://devcenter.bitrise.io/en/builds/build-cache.html'
            },
            'duplicates': {
                icon: '🔄',
                title: 'Duplicate Files',
                learnMore: 'https://devcenter.bitrise.io/en/builds/build-cache.html'
            },
            'image-optimization': {
                icon: '🖼️',
                title: 'Image Optimization',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/optimizing-app-size.html'
            },
            'loose-images': {
                icon: '📸',
                title: 'Loose Images',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/optimizing-app-size.html'
            },
            'unnecessary-files': {
                icon: '🗑️',
                title: 'Unnecessary Files',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/optimizing-app-size.html'
            }
        };

        // Group optimizations by category
        function groupByCategory(optimizations) {
            const groups = {};

            optimizations.forEach(opt => {
                if (!groups[opt.category]) {
                    groups[opt.category] = {
                        items: [],
                        totalSavings: 0,
                        totalFiles: 0,
                        description: ''
                    };
                }

                groups[opt.category].items.push(opt);
                groups[opt.category].totalSavings += opt.impact;
                groups[opt.category].totalFiles += (opt.files ? opt.files.length : 0);

                // Use the first item's description as category description
                if (!groups[opt.category].description && opt.description) {
                    groups[opt.category].description = opt.description;
                }
            });

            return groups;
        }

        // Render insights section
        function renderInsights(optimizations) {
            const container = document.getElementById('insights-list');

            if (!optimizations || optimizations.length === 0) {
                container.innerHTML = '<div class="no-insights">✅ No optimization opportunities found! Your bundle is well optimized.</div>';
                return;
            }

            const groups = groupByCategory(optimizations);

            // Calculate total bundle size for percentage
            const totalSize = reportData.fileTree ? reportData.fileTree.value : 0;

            let html = '';

            // Render each category
            Object.keys(groups).forEach((category, index) => {
                const group = groups[category];
                const metadata = categoryMetadata[category] || {
                    icon: '💡',
                    title: category.replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase()),
                    learnMore: 'https://devcenter.bitrise.io'
                };

                const savingsPercentage = totalSize > 0
                    ? ((group.totalSavings / totalSize) * 100).toFixed(2)
                    : '0.00';

                html += '<div class="insight-card" id="insight-' + index + '">';
                html += '  <div class="insight-header" onclick="toggleInsight(' + index + ')">';
                html += '    <div class="insight-icon">' + metadata.icon + '</div>';
                html += '    <div class="insight-content">';
                html += '      <div class="insight-title">' + metadata.title + '</div>';
                html += '      <div class="insight-description">' + group.description + '</div>';
                html += '      <div class="insight-meta">';
                html += '        <div class="insight-savings">';
                html += '          <span class="savings-amount">' + formatBytes(group.totalSavings) + '</span>';
                html += '          <span class="savings-percentage">' + savingsPercentage + '% of total</span>';
                html += '        </div>';
                html += '        <span class="insight-count">' + group.totalFiles + ' files</span>';
                html += '        <a href="' + metadata.learnMore + '" class="learn-more-link" target="_blank" onclick="event.stopPropagation()">Learn more →</a>';
                html += '      </div>';
                html += '    </div>';
                html += '    <div class="expand-indicator">▼</div>';
                html += '  </div>';
                html += '  <div class="insight-files">';
                html += '    <div class="insight-files-content">';
                html += '      <div class="insight-files-header">Affected Files</div>';
                html += '      <ul class="files-list">';

                // Collect all unique files from all items in this category
                const allFiles = new Set();
                group.items.forEach(item => {
                    if (item.files) {
                        item.files.forEach(file => allFiles.add(file));
                    }
                });

                Array.from(allFiles).forEach(file => {
                    html += '<li>' + file + '</li>';
                });

                html += '      </ul>';
                html += '    </div>';
                html += '  </div>';
                html += '</div>';
            });

            container.innerHTML = html;
        }

        // Toggle insight card expansion
        function toggleInsight(index) {
            const card = document.getElementById('insight-' + index);
            card.classList.toggle('expanded');
        }

        // Initialize visualizations
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize theme first
            initTheme();

            // Create charts and store instances
            if (reportData.fileTree) {
                treemapChart = createTreemap(reportData.fileTree);
            }
            if (reportData.categories && reportData.categories.length > 0) {
                categoryChart = createCategoryChart(reportData.categories);
            }
            if (reportData.extensions && reportData.extensions.length > 0) {
                extensionChart = createExtensionChart(reportData.extensions);
            }
            if (reportData.optimizations) {
                renderInsights(reportData.optimizations);
            }
        });

        // Search functionality (basic implementation)
        document.getElementById('search-input').addEventListener('input', function(e) {
            const query = e.target.value.toLowerCase();
            // TODO: Implement search highlighting in treemap
            console.log('Search:', query);
        });
    </script>
</body>
</html>
`
