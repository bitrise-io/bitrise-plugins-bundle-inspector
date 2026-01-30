package report

// htmlTemplate is the embedded HTML template for the interactive report
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
            background: #f5f5f7;
            color: #1d1d1f;
            line-height: 1.6;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            padding: 20px;
        }

        header {
            background: white;
            border-radius: 12px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        h1 {
            font-size: 32px;
            font-weight: 600;
            margin-bottom: 10px;
        }

        .artifact-info {
            display: flex;
            gap: 15px;
            flex-wrap: wrap;
            margin-top: 15px;
            color: #6e6e73;
        }

        .artifact-info span {
            display: flex;
            align-items: center;
            gap: 5px;
        }

        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }

        .metric-card {
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .metric-label {
            font-size: 14px;
            color: #6e6e73;
            margin-bottom: 8px;
        }

        .metric-value {
            font-size: 28px;
            font-weight: 600;
            color: #1d1d1f;
        }

        .metric-secondary {
            font-size: 14px;
            color: #86868b;
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
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            min-height: 600px;
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
            background: white;
            border-radius: 12px;
            padding: 20px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .chart-card h2 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 15px;
        }

        .chart {
            width: 100%;
            height: 280px;
        }

        .optimizations-section {
            background: white;
            border-radius: 12px;
            padding: 30px;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        }

        .optimizations-section h2 {
            font-size: 24px;
            font-weight: 600;
            margin-bottom: 20px;
        }

        .optimization-item {
            border: 1px solid #e5e5e7;
            border-radius: 8px;
            padding: 20px;
            margin-bottom: 15px;
        }

        .optimization-item:last-child {
            margin-bottom: 0;
        }

        .opt-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 10px;
        }

        .badge {
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: 600;
            text-transform: uppercase;
        }

        .badge.high {
            background: #ff3b30;
            color: white;
        }

        .badge.medium {
            background: #ff9500;
            color: white;
        }

        .badge.low {
            background: #ffcc00;
            color: #1d1d1f;
        }

        .opt-header h3 {
            flex: 1;
            font-size: 16px;
            font-weight: 600;
        }

        .impact {
            font-size: 14px;
            font-weight: 600;
            color: #34c759;
        }

        .description {
            color: #6e6e73;
            margin-bottom: 10px;
        }

        details {
            margin-top: 10px;
        }

        summary {
            cursor: pointer;
            color: #007aff;
            font-weight: 500;
            user-select: none;
        }

        summary:hover {
            text-decoration: underline;
        }

        .files-list {
            list-style: none;
            margin-top: 10px;
            max-height: 200px;
            overflow-y: auto;
        }

        .files-list li {
            padding: 4px 0;
            color: #6e6e73;
            font-family: 'Monaco', 'Courier New', monospace;
            font-size: 12px;
        }

        .action {
            margin-top: 10px;
            padding: 10px;
            background: #f5f5f7;
            border-radius: 6px;
            font-size: 14px;
        }

        .action strong {
            color: #007aff;
        }

        .search-container {
            margin-bottom: 15px;
        }

        #search-input {
            width: 100%%;
            padding: 10px 15px;
            border: 1px solid #d2d2d7;
            border-radius: 8px;
            font-size: 14px;
            outline: none;
            transition: border-color 0.2s;
        }

        #search-input:focus {
            border-color: #007aff;
        }

        .legend {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            margin-top: 15px;
            padding-top: 15px;
            border-top: 1px solid #e5e5e7;
        }

        .legend-item {
            display: flex;
            align-items: center;
            gap: 5px;
            font-size: 12px;
        }

        .legend-color {
            width: 16px;
            height: 16px;
            border-radius: 3px;
        }

        footer {
            text-align: center;
            padding: 20px;
            color: #86868b;
            font-size: 14px;
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
            <h1>{{.Title}}</h1>
            <div class="artifact-info">
                <span><strong>Artifact:</strong> {{.ArtifactName}}</span>
                <span><strong>Type:</strong> {{.ArtifactType}}</span>
                <span><strong>Analyzed:</strong> <time>{{.Timestamp}}</time></span>
            </div>
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

        <div class="optimizations-section" id="optimizations-section">
            <h2>Optimization Opportunities</h2>
            <div id="optimizations-list"></div>
        </div>

        <footer>
            Generated by Bundle Inspector | <a href="https://github.com/bitrise-io/bitrise-plugins-bundle-inspector" target="_blank">GitHub</a>
        </footer>
    </div>

    <script src="https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js"></script>
    <script>
        const reportData = {{.DataJSON}};

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

        // Create treemap visualization
        function createTreemap(data) {
            const container = document.getElementById('treemap');
            const myChart = echarts.init(container, null, {
                renderer: 'canvas',
                useDirtyRect: true
            });

            // Apply colors
            applyColorsToTree(data);

            const option = {
                tooltip: {
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
                            color: 'rgba(255,255,255,0.7)',
                            borderColor: 'rgba(255,255,255,0.7)',
                            textStyle: {
                                color: '#333'
                            }
                        },
                        emphasis: {
                            itemStyle: {
                                color: 'rgba(255,255,255,1)',
                                textStyle: {
                                    color: '#000'
                                }
                            }
                        }
                    },
                    label: {
                        show: true,
                        formatter: '{b}',
                        fontSize: 11,
                        overflow: 'truncate'
                    },
                    upperLabel: {
                        show: true,
                        height: 25,
                        color: '#fff',
                        textBorderColor: 'transparent'
                    },
                    itemStyle: {
                        borderColor: '#fff',
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
                            borderColor: '#333',
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

            myChart.setOption(option);

            window.addEventListener('resize', function() {
                myChart.resize();
            });

            return myChart;
        }

        // Create category donut chart
        function createCategoryChart(categories) {
            const container = document.getElementById('category-chart');
            const myChart = echarts.init(container);

            const option = {
                tooltip: {
                    trigger: 'item',
                    formatter: '{b}: {c} ({d}%%)'
                },
                series: [{
                    type: 'pie',
                    radius: ['40%%', '70%%'],
                    avoidLabelOverlap: true,
                    itemStyle: {
                        borderRadius: 8,
                        borderColor: '#fff',
                        borderWidth: 2
                    },
                    label: {
                        show: true,
                        formatter: function(params) {
                            return params.name + '\n' + formatBytes(params.value);
                        },
                        fontSize: 11
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

            myChart.setOption(option);
            window.addEventListener('resize', () => myChart.resize());
        }

        // Create extension bar chart
        function createExtensionChart(extensions) {
            const container = document.getElementById('extension-chart');
            const myChart = echarts.init(container);

            const option = {
                tooltip: {
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' },
                    formatter: function(params) {
                        const data = params[0];
                        return data.name + ': ' + formatBytes(data.value);
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
                        fontSize: 10
                    }
                },
                yAxis: {
                    type: 'category',
                    data: extensions.map(e => e.name),
                    axisLabel: {
                        fontSize: 10
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
                        fontSize: 10
                    }
                }]
            };

            myChart.setOption(option);
            window.addEventListener('resize', () => myChart.resize());
        }

        // Render optimizations
        function renderOptimizations(optimizations) {
            const container = document.getElementById('optimizations-list');

            if (!optimizations || optimizations.length === 0) {
                container.innerHTML = '<p style="color: #34c759; font-size: 18px;">✅ No optimization opportunities found!</p>';
                return;
            }

            let html = '';
            optimizations.forEach(opt => {
                html += '<div class="optimization-item">';
                html += '<div class="opt-header">';
                html += '<span class="badge ' + opt.severity + '">' + opt.severity + '</span>';
                html += '<h3>' + opt.title + '</h3>';
                html += '<span class="impact">' + formatBytes(opt.impact) + '</span>';
                html += '</div>';
                html += '<p class="description">' + opt.description + '</p>';

                if (opt.files && opt.files.length > 0) {
                    html += '<details>';
                    html += '<summary>Affected Files (' + opt.files.length + ')</summary>';
                    html += '<ul class="files-list">';
                    opt.files.forEach(file => {
                        html += '<li>' + file + '</li>';
                    });
                    html += '</ul>';
                    html += '</details>';
                }

                html += '<div class="action"><strong>Action:</strong> ' + opt.action + '</div>';
                html += '</div>';
            });

            container.innerHTML = html;
        }

        // Initialize visualizations
        document.addEventListener('DOMContentLoaded', function() {
            if (reportData.fileTree) {
                createTreemap(reportData.fileTree);
            }
            if (reportData.categories && reportData.categories.length > 0) {
                createCategoryChart(reportData.categories);
            }
            if (reportData.extensions && reportData.extensions.length > 0) {
                createExtensionChart(reportData.extensions);
            }
            if (reportData.optimizations) {
                renderOptimizations(reportData.optimizations);
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
