package report

// htmlTemplate is the embedded HTML template for the interactive report
const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Lato:wght@300;400;600;700;900&family=IBM+Plex+Mono:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <script src="https://cdn.tailwindcss.com"></script>
    <script>
        tailwind.config = {
            darkMode: 'class',
            theme: {
                extend: {
                    colors: {
                        border: "hsl(var(--border))",
                        input: "hsl(var(--input))",
                        ring: "hsl(var(--ring))",
                        background: "hsl(var(--background))",
                        foreground: "hsl(var(--foreground))",
                        primary: {
                            DEFAULT: "#9247C2",
                            dark: "#351d48",
                            foreground: "#ffffff",
                        },
                        secondary: {
                            DEFAULT: "#0dd3c5",
                            foreground: "#000000",
                        },
                        success: {
                            DEFAULT: "#34c759",
                            foreground: "#ffffff",
                        },
                        muted: {
                            DEFAULT: "hsl(var(--muted))",
                            foreground: "hsl(var(--muted-foreground))",
                        },
                        accent: {
                            DEFAULT: "hsl(var(--accent))",
                            foreground: "hsl(var(--accent-foreground))",
                        },
                        card: {
                            DEFAULT: "hsl(var(--card))",
                            foreground: "hsl(var(--card-foreground))",
                        },
                    },
                    borderRadius: {
                        lg: "var(--radius)",
                        md: "calc(var(--radius) - 2px)",
                        sm: "calc(var(--radius) - 4px)",
                    },
                    fontFamily: {
                        sans: ['Lato', 'system-ui', 'sans-serif'],
                        mono: ['IBM Plex Mono', 'monospace'],
                    },
                }
            }
        }
    </script>
    <style>
        :root {
            /* shadcn/ui style variables - Light mode */
            --background: 210 40% 98%;
            --foreground: 222.2 84% 4.9%;
            --card: 0 0% 100%;
            --card-foreground: 222.2 84% 4.9%;
            --muted: 210 40% 96.1%;
            --muted-foreground: 215.4 16.3% 46.9%;
            --accent: 210 40% 96.1%;
            --accent-foreground: 222.2 47.4% 11.2%;
            --border: 214.3 31.8% 91.4%;
            --input: 214.3 31.8% 91.4%;
            --ring: 274 58% 52%;
            --radius: 0.75rem;

            /* File Type Colors */
            --color-framework: #9247C2;
            --color-library: #0dd3c5;
            --color-native: #ff9500;
            --color-image: #ffd60a;
            --color-asset-catalog: #30d158;
            --color-resource: #64d2ff;
            --color-ui: #bf5af2;
            --color-dex: #ac8e68;
            --color-duplicate: #ff453a;
            --color-other: #98989d;
        }

        .dark {
            /* shadcn/ui style variables - Dark mode */
            --background: 222.2 84% 4.9%;
            --foreground: 210 40% 98%;
            --card: 222.2 84% 8%;
            --card-foreground: 210 40% 98%;
            --muted: 217.2 32.6% 17.5%;
            --muted-foreground: 215 20.2% 65.1%;
            --accent: 217.2 32.6% 17.5%;
            --accent-foreground: 210 40% 98%;
            --border: 217.2 32.6% 17.5%;
            --input: 217.2 32.6% 17.5%;
            --ring: 274 58% 52%;
        }

        @layer base {
            * {
                @apply border-border;
            }
            body {
                @apply bg-background text-foreground;
                font-feature-settings: "rlig" 1, "calt" 1;
            }
        }

        /* Chart specific */
        .chart {
            width: 100%;
            height: 450px;
        }

        #treemap {
            width: 100%;
            height: 600px;
        }

        /* Tab animations */
        .tab-panel {
            display: none;
        }

        .tab-panel.active {
            display: block;
            animation: fadeIn 0.3s ease;
        }

        @keyframes fadeIn {
            from {
                opacity: 0;
                transform: translateY(10px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }

        /* Insight card accordion animation */
        .insight-files {
            max-height: 0;
            overflow: hidden;
            transition: max-height 0.3s ease;
        }

        .insight-card.expanded .insight-files {
            overflow-y: auto;
        }

        .expand-indicator {
            transition: transform 0.3s ease;
        }

        .insight-card.expanded .expand-indicator {
            transform: rotate(180deg);
        }
    </style>
</head>
<body class="font-sans antialiased bg-background text-foreground">
    <!-- Top Navigation Bar -->
    <header class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div class="w-full max-w-7xl mx-auto px-6">
            <div class="flex h-16 items-center justify-between">
                <!-- Bitrise Logo -->
                <div class="flex items-center gap-3">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 240 240" class="h-8 w-8" aria-label="Bitrise">
                        <defs>
                            <linearGradient id="bitrise-gradient" x1="0" y1="0" x2="1" y2="1">
                                <stop offset="0" style="stop-color:#9247C2;stop-opacity:1" />
                                <stop offset="1" style="stop-color:#0dd3c5;stop-opacity:1" />
                            </linearGradient>
                        </defs>
                        <rect x="20" y="20" width="200" height="200" rx="40" fill="url(#bitrise-gradient)"/>
                        <path d="M120 80 L160 120 L120 160 L80 120 Z" fill="white" opacity="0.9"/>
                        <circle cx="120" cy="120" r="15" fill="white"/>
                    </svg>
                    <div class="flex flex-col">
                        <span class="font-bold text-lg leading-tight">Bundle Inspector</span>
                        <span class="text-xs text-muted-foreground">Size Analysis Report</span>
                    </div>
                </div>

                <!-- Dark Mode Toggle with Keyboard Shortcut -->
                <button onclick="toggleTheme()"
                        class="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-9 w-9 relative group"
                        aria-label="Toggle theme (Press D)"
                        title="Toggle theme (Press D)">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="h-4.5 w-4.5 transition-transform group-hover:scale-110">
                        <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                        <path d="M12 12m-9 0a9 9 0 1 0 18 0a9 9 0 1 0 -18 0"></path>
                        <path d="M12 3l0 18"></path>
                        <path d="M12 9l4.65 -4.65"></path>
                        <path d="M12 14.3l7.37 -7.37"></path>
                        <path d="M12 19.6l8.85 -8.85"></path>
                    </svg>
                    <kbd class="pointer-events-none absolute -bottom-8 left-1/2 -translate-x-1/2 hidden group-hover:inline-flex h-5 select-none items-center gap-1 rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground opacity-100 shadow-sm">
                        D
                    </kbd>
                </button>
            </div>
        </div>
    </header>

    <!-- Main Content Area -->
    <main class="w-full">
        <div class="w-full max-w-7xl mx-auto px-6 py-6 space-y-6">
        <!-- App Info Card -->
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6">
            <h1 class="text-3xl font-bold tracking-tight mb-4">{{if .AppName}}{{.AppName}}{{else}}{{.Title}}{{end}}</h1>
            <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mt-4">
                        <!-- App Info -->
                        <div class="space-y-4">
                            <h3 class="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-3">App Info</h3>
                            <div class="space-y-3">
                                {{if .BundleID}}<div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Bundle ID</span>
                                    <span class="text-sm font-semibold text-right">{{.BundleID}}</span>
                                </div>{{end}}
                                {{if .Platform}}<div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Platform</span>
                                    <span class="text-sm font-semibold text-right">{{.Platform}}</span>
                                </div>{{end}}
                                {{if .Version}}<div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Version</span>
                                    <span class="text-sm font-semibold text-right">{{.Version}}</span>
                                </div>{{end}}
                            </div>
                        </div>
                        <!-- Build Info -->
                        <div class="space-y-4">
                            <h3 class="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-3">Build Info</h3>
                            <div class="space-y-3">
                                {{if .Branch}}<div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Branch</span>
                                    <span class="text-sm font-semibold text-right">{{.Branch}}</span>
                                </div>{{end}}
                                {{if .CommitSHA}}<div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Commit</span>
                                    <span class="text-sm font-semibold font-mono text-right bg-muted px-2 py-0.5 rounded">{{.CommitSHA}}</span>
                                </div>{{end}}
                                <div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Analyzed</span>
                                    <span class="text-sm font-semibold text-right"><time>{{.Timestamp}}</time></span>
                                </div>
                            </div>
                        </div>
                        <!-- Size Analysis -->
                        <div class="space-y-4">
                            <h3 class="text-xs font-bold uppercase tracking-wider text-muted-foreground mb-3">Size Analysis</h3>
                            <div class="space-y-3">
                                <div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Download Size</span>
                                    <span class="text-sm font-semibold text-right">{{.TotalSize}}</span>
                                </div>
                                <div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Install Size</span>
                                    <span class="text-sm font-semibold text-right">{{.UncompressedSize}}</span>
                                </div>
                                <div class="flex justify-between items-baseline gap-4">
                                    <span class="text-sm text-muted-foreground font-medium">Potential Savings</span>
                                    <span class="inline-flex items-center gap-1 text-sm font-semibold bg-success/10 text-success px-2.5 py-1 rounded-md">{{.TotalSavings}}</span>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

            <!-- Tabs -->
            <div class="space-y-4">
            <div class="inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground">
                <button class="tab-button active inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm"
                        onclick="switchTab('app-analyzer')">App Analyzer</button>
                <button class="tab-button inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
                        onclick="switchTab('category')">Category</button>
            </div>

            <div id="app-analyzer-panel" class="tab-panel active">
                <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6 hover:-translate-y-0.5 transition-transform duration-300">
                    <h2 class="text-2xl font-bold tracking-tight mb-2">Bundle Treemap</h2>
                    <p class="text-sm text-muted-foreground mb-4">Click to drill down into folders. Use mouse wheel to zoom. Use breadcrumb to navigate back.</p>
                    <div class="mb-4">
                        <div class="relative">
                            <svg class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                            </svg>
                            <input type="text" id="search-input"
                                   class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                   placeholder="Search files (e.g., .png, Frameworks/, &#96;Assets.car&#96;)">
                        </div>
                    </div>
                    <div id="treemap"></div>
                    <div class="flex flex-wrap gap-4 mt-4 pt-4 border-t border-border">
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-duplicate);"></div>
                            <span class="text-xs text-muted-foreground">Duplicates</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-framework);"></div>
                            <span class="text-xs text-muted-foreground">Frameworks</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-library);"></div>
                            <span class="text-xs text-muted-foreground">Libraries</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-image);"></div>
                            <span class="text-xs text-muted-foreground">Images</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-dex);"></div>
                            <span class="text-xs text-muted-foreground">DEX</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-native);"></div>
                            <span class="text-xs text-muted-foreground">Native Libs</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-asset-catalog);"></div>
                            <span class="text-xs text-muted-foreground">Asset Catalogs</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-resource);"></div>
                            <span class="text-xs text-muted-foreground">Resources</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-ui);"></div>
                            <span class="text-xs text-muted-foreground">UI</span>
                        </div>
                        <div class="flex items-center gap-2">
                            <div class="w-4 h-4 rounded" style="background: var(--color-other);"></div>
                            <span class="text-xs text-muted-foreground">Other</span>
                        </div>
                    </div>
                </div>
            </div>

            <div id="category-panel" class="tab-panel">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6 hover:-translate-y-0.5 transition-transform duration-300">
                        <h2 class="text-xl font-bold tracking-tight mb-6">Category Breakdown</h2>
                        <div id="category-chart" class="chart"></div>
                    </div>
                    <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6 hover:-translate-y-0.5 transition-transform duration-300">
                        <h2 class="text-xl font-bold tracking-tight mb-6">Top Extensions</h2>
                        <div id="extension-chart" class="chart"></div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Insights Section -->
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6" id="insights-section">
            <h2 class="text-2xl font-bold tracking-tight mb-6">💡 Insights & Optimization Opportunities</h2>
            <div id="insights-list" class="space-y-4"></div>
        </div>

        </div>
    </main>

    <!-- Footer -->
    <footer class="w-full border-t">
        <div class="w-full max-w-7xl mx-auto px-6 py-6">
            <p class="text-center text-sm text-muted-foreground">
                Generated by Bundle Inspector | <a href="https://github.com/bitrise-io/bitrise-plugins-bundle-inspector" target="_blank" class="font-semibold text-primary hover:underline transition-colors">GitHub</a>
            </p>
        </div>
    </footer>

    <script src="https://cdn.jsdelivr.net/npm/echarts@5/dist/echarts.min.js"></script>
    <script>
        const reportData = {{.DataJSON}};

        // Theme management
        let currentTheme = 'light';
        let treemapChart = null;
        let categoryChart = null;
        let extensionChart = null;

        // Store original data for search filtering
        let originalFileTree = null;
        let originalCategories = null;
        let originalExtensions = null;

        // Tab switching functionality
        function switchTab(tabName) {
            // Remove active class and styles from all tabs and panels
            const tabButtons = document.querySelectorAll('.tab-button');
            const tabPanels = document.querySelectorAll('.tab-panel');

            tabButtons.forEach(button => {
                button.classList.remove('active');
                button.classList.remove('bg-background', 'text-foreground', 'shadow-sm');
            });
            tabPanels.forEach(panel => panel.classList.remove('active'));

            // Add active class and styles to selected tab and panel
            const selectedButton = event.target;
            selectedButton.classList.add('active', 'bg-background', 'text-foreground', 'shadow-sm');

            const selectedPanel = document.getElementById(tabName + '-panel');
            selectedPanel.classList.add('active');

            // Resize charts when switching to category tab
            if (tabName === 'category') {
                setTimeout(() => {
                    if (categoryChart) categoryChart.resize();
                    if (extensionChart) extensionChart.resize();
                }, 100);
            } else if (tabName === 'app-analyzer') {
                setTimeout(() => {
                    if (treemapChart) treemapChart.resize();
                }, 100);
            }
        }

        // Initialize theme from localStorage
        function initTheme() {
            const savedTheme = localStorage.getItem('theme') || 'light';
            currentTheme = savedTheme;
            if (savedTheme === 'dark') {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
            updateThemeButton();
        }

        // Toggle theme
        function toggleTheme() {
            currentTheme = currentTheme === 'light' ? 'dark' : 'light';
            if (currentTheme === 'dark') {
                document.documentElement.classList.add('dark');
            } else {
                document.documentElement.classList.remove('dark');
            }
            localStorage.setItem('theme', currentTheme);
            updateThemeButton();
            updateChartsTheme();
        }

        // Update theme button icon (rotate for visual feedback)
        function updateThemeButton() {
            const buttons = document.querySelectorAll('[onclick="toggleTheme()"]');
            buttons.forEach(button => {
                const svg = button.querySelector('svg');
                if (svg) {
                    if (currentTheme === 'dark') {
                        // Rotate icon for dark mode
                        svg.style.transform = 'rotate(180deg)';
                        button.setAttribute('aria-label', 'Switch to light mode (Press D)');
                        button.setAttribute('title', 'Switch to light mode (Press D)');
                    } else {
                        // Normal orientation for light mode
                        svg.style.transform = 'rotate(0deg)';
                        button.setAttribute('aria-label', 'Switch to dark mode (Press D)');
                        button.setAttribute('title', 'Switch to dark mode (Press D)');
                    }
                }
            });
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

        // Truncate path while preserving filename
        function truncatePath(path, maxLen) {
            maxLen = maxLen || 80;
            if (path.length <= maxLen) {
                return path;
            }

            // Find the last path separator to identify the filename
            const lastSep = path.lastIndexOf('/');
            if (lastSep === -1) {
                // No separator, just truncate from beginning
                if (maxLen > 3) {
                    return '...' + path.slice(-(maxLen - 3));
                }
                return path.slice(0, maxLen);
            }

            const filename = path.slice(lastSep + 1);
            const dirPath = path.slice(0, lastSep);

            // If filename alone is longer than maxLen, truncate it
            if (filename.length >= maxLen - 4) { // 4 for ".../"
                return '.../' + filename.slice(0, maxLen - 4);
            }

            // Calculate how much of the directory path we can keep
            const ellipsis = '/.../';
            const availableForDir = maxLen - filename.length - ellipsis.length;

            if (availableForDir <= 0) {
                // Just show the filename with ellipsis
                return '.../' + filename;
            }

            // Take characters from the start of the directory path
            if (availableForDir >= dirPath.length) {
                return path; // Shouldn't happen, but safety check
            }

            return dirPath.slice(0, availableForDir) + ellipsis + filename;
        }

        // Get CSS variable value
        function getCSSVariable(name) {
            return getComputedStyle(document.documentElement).getPropertyValue(name).trim();
        }

        // Color mapping for file types - using CSS variables
        function getFileTypeColors() {
            return {
                'framework': getCSSVariable('--color-framework'),
                'library': getCSSVariable('--color-library'),
                'native': getCSSVariable('--color-native'),
                'image': getCSSVariable('--color-image'),
                'asset_catalog': getCSSVariable('--color-asset-catalog'),
                'resource': getCSSVariable('--color-resource'),
                'ui': getCSSVariable('--color-ui'),
                'dex': getCSSVariable('--color-dex'),
                'font': getCSSVariable('--color-error'),
                'other': getCSSVariable('--color-other'),
                'duplicate': getCSSVariable('--color-duplicate')
            };
        }

        // Create a Set of duplicate file paths for fast lookup
        const duplicatePaths = new Set(reportData.duplicates || []);

        // Get color for file type
        function getColorForFileType(fileType) {
            const colors = getFileTypeColors();
            return colors[fileType] || colors['other'];
        }

        // Darken a hex color by a factor (0 = original, 1 = black)
        function darkenColor(hex, factor) {
            // Remove # if present
            hex = hex.replace(/^#/, '');

            // Parse RGB
            let r = parseInt(hex.substring(0, 2), 16);
            let g = parseInt(hex.substring(2, 4), 16);
            let b = parseInt(hex.substring(4, 6), 16);

            // Darken
            r = Math.round(r * (1 - factor));
            g = Math.round(g * (1 - factor));
            b = Math.round(b * (1 - factor));

            // Convert back to hex
            return '#' + [r, g, b].map(x => x.toString(16).padStart(2, '0')).join('');
        }

        // Get the dominant file type from a node's descendants
        function getDominantFileType(node) {
            if (node.fileType) {
                return node.fileType;
            }
            if (!node.children || node.children.length === 0) {
                return 'other';
            }

            // Count file types by total size
            const typeSizes = {};
            function countTypes(n) {
                if (n.fileType && n.value) {
                    typeSizes[n.fileType] = (typeSizes[n.fileType] || 0) + n.value;
                }
                if (n.children) {
                    n.children.forEach(countTypes);
                }
            }
            countTypes(node);

            // Find the type with the largest total size
            let dominantType = 'other';
            let maxSize = 0;
            for (const type in typeSizes) {
                if (typeSizes[type] > maxSize) {
                    maxSize = typeSizes[type];
                    dominantType = type;
                }
            }
            return dominantType;
        }

        // Apply colors to tree nodes with depth-based darkening
        function applyColorsToTree(node, depth) {
            depth = depth || 0;
            const isParent = node.children && node.children.length > 0;

            // Parent nodes (folders) get darker colors for readable headers
            // Leaf nodes get progressively lighter colors based on depth
            let darkenFactor;
            if (isParent) {
                // Parents are dark (0.4 base) and get slightly darker with depth
                darkenFactor = 0.4 + Math.min(depth * 0.05, 0.15);
            } else {
                // Leaves are lighter and vary with depth
                darkenFactor = Math.min(depth * 0.08, 0.3);
            }

            // Determine the base color for this node
            let baseColor;
            const colors = getFileTypeColors();
            if (node.path && duplicatePaths.has(node.path)) {
                baseColor = colors['duplicate'];
                node.isDuplicate = true;
            } else {
                const fileType = node.fileType || getDominantFileType(node);
                baseColor = getColorForFileType(fileType);
            }

            // Apply color with darkening
            const finalColor = darkenColor(baseColor, darkenFactor);
            node.itemStyle = node.itemStyle || {};
            node.itemStyle.color = finalColor;

            // For parent nodes, also set a darker border to help define the header area
            if (isParent) {
                node.itemStyle.borderColor = darkenColor(baseColor, darkenFactor + 0.2);
            }
            if (node.children) {
                node.children.forEach(child => applyColorsToTree(child, depth + 1));
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
                        const isDuplicate = info.data.isDuplicate || false;
                        const treePathInfo = info.treePathInfo || [];
                        let percentage = '0.0';

                        if (treePathInfo.length > 0) {
                            const rootValue = treePathInfo[0].value;
                            if (rootValue > 0) {
                                percentage = ((value / rootValue) * 100).toFixed(2);
                            }
                        }

                        let result = '<strong>' + name + '</strong><br/>' +
                               'Path: ' + path + '<br/>' +
                               'Size: ' + formatBytes(value) + '<br/>' +
                               percentage + '%% of total';

                        if (isDuplicate) {
                            result += '<br/><span style="color: #e74c3c; font-weight: bold;">⚠ Duplicate file</span>';
                        }

                        return result;
                    }
                },
                series: [{
                    type: 'treemap',
                    width: '100%%',
                    height: '100%%',
                    roam: true,
                    nodeClick: 'zoomToNode',
                    leafDepth: 4,
                    zoomToNodeRatio: 0.32 * 0.32,
                    scaleLimit: {
                        min: 0.5,
                        max: 20
                    },
                    drillDownIcon: '▶',
                    colorMappingBy: 'value',
                    breadcrumb: {
                        show: true,
                        top: 5,
                        left: 5,
                        height: 28,
                        emptyItemWidth: 25,
                        itemStyle: {
                            color: isDark ? 'rgba(60,60,60,0.95)' : 'rgba(80,80,80,0.9)',
                            borderColor: isDark ? 'rgba(80,80,80,0.9)' : 'rgba(60,60,60,0.8)',
                            borderWidth: 1,
                            borderRadius: 4,
                            textStyle: {
                                color: '#fff',
                                fontSize: 12
                            }
                        },
                        emphasis: {
                            itemStyle: {
                                color: isDark ? 'rgba(80,80,80,1)' : 'rgba(60,60,60,1)',
                                textStyle: {
                                    color: '#fff'
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
                    visibleMin: 200,
                    childrenVisibleMin: 100,
                    levels: [
                        {
                            // Level 0: Root level
                            itemStyle: {
                                borderWidth: 0,
                                gapWidth: 4
                            },
                            upperLabel: {
                                show: false
                            }
                        },
                        {
                            // Level 1: Main categories
                            itemStyle: {
                                gapWidth: 2,
                                borderWidth: 2,
                                borderColor: isDark ? '#444' : '#ddd'
                            },
                            upperLabel: {
                                show: true,
                                height: 28,
                                formatter: function(params) {
                                    return '{bg|' + params.name + '}';
                                },
                                rich: {
                                    bg: {
                                        backgroundColor: isDark ? 'rgba(0,0,0,0.6)' : 'rgba(0,0,0,0.7)',
                                        color: '#fff',
                                        fontWeight: 'bold',
                                        fontSize: 13,
                                        padding: [4, 8],
                                        borderRadius: 3
                                    }
                                }
                            }
                        },
                        {
                            // Level 2
                            itemStyle: {
                                gapWidth: 2,
                                borderWidth: 1,
                                borderColor: isDark ? '#555' : '#eee'
                            },
                            upperLabel: {
                                show: true,
                                height: 24,
                                formatter: function(params) {
                                    return '{bg|' + params.name + '}';
                                },
                                rich: {
                                    bg: {
                                        backgroundColor: isDark ? 'rgba(0,0,0,0.5)' : 'rgba(0,0,0,0.6)',
                                        color: '#fff',
                                        fontWeight: 'bold',
                                        fontSize: 12,
                                        padding: [3, 6],
                                        borderRadius: 3
                                    }
                                }
                            }
                        },
                        {
                            // Level 3
                            itemStyle: {
                                gapWidth: 1,
                                borderWidth: 1,
                                borderColor: isDark ? '#555' : '#eee'
                            },
                            upperLabel: {
                                show: true,
                                height: 22,
                                formatter: function(params) {
                                    return '{bg|' + params.name + '}';
                                },
                                rich: {
                                    bg: {
                                        backgroundColor: isDark ? 'rgba(0,0,0,0.4)' : 'rgba(0,0,0,0.5)',
                                        color: '#fff',
                                        fontSize: 11,
                                        padding: [2, 5],
                                        borderRadius: 2
                                    }
                                }
                            }
                        },
                        {
                            // Level 4+
                            itemStyle: {
                                gapWidth: 1,
                                borderWidth: 1
                            },
                            upperLabel: {
                                show: true,
                                height: 20,
                                formatter: function(params) {
                                    return '{bg|' + params.name + '}';
                                },
                                rich: {
                                    bg: {
                                        backgroundColor: isDark ? 'rgba(0,0,0,0.35)' : 'rgba(0,0,0,0.45)',
                                        color: '#fff',
                                        fontSize: 10,
                                        padding: [2, 4],
                                        borderRadius: 2
                                    }
                                }
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
                container.innerHTML = '<div class="text-center py-10 text-lg font-semibold text-success">✅ No optimization opportunities found! Your bundle is well optimized.</div>';
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

                html += '<div class="insight-card rounded-lg border bg-card overflow-hidden transition-all duration-200 hover:shadow-md" id="insight-' + index + '">';
                html += '  <div class="flex items-start gap-3 p-4 cursor-pointer select-none" onclick="toggleInsight(' + index + ')">';
                html += '    <div class="text-2xl flex-shrink-0 leading-none mt-0.5">' + metadata.icon + '</div>';
                html += '    <div class="flex-1 min-w-0">';
                html += '      <div class="flex items-center justify-between gap-2 mb-1.5">';
                html += '        <h3 class="text-base font-semibold leading-none">' + metadata.title + '</h3>';
                html += '        <div class="expand-indicator text-base text-muted-foreground flex-shrink-0 transition-transform duration-200">▼</div>';
                html += '      </div>';
                html += '      <p class="text-sm text-muted-foreground mb-2.5 leading-snug">' + group.description + '</p>';
                html += '      <div class="flex items-center gap-2 flex-wrap text-xs">';
                html += '        <span class="inline-flex items-center gap-1.5 font-semibold bg-success/10 text-success px-2 py-1 rounded-md">';
                html += '          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>';
                html += '          ' + formatBytes(group.totalSavings) + ' (' + savingsPercentage + '%)';
                html += '        </span>';
                html += '        <span class="inline-flex items-center gap-1 text-muted-foreground px-2 py-1 bg-muted/50 rounded-md font-medium">';
                html += '          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>';
                html += '          ' + group.totalFiles + ' files';
                html += '        </span>';
                html += '        <a href="' + metadata.learnMore + '" class="inline-flex items-center gap-0.5 text-primary hover:text-primary/80 font-semibold transition-colors" target="_blank" onclick="event.stopPropagation()">';
                html += '          Learn more';
                html += '          <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>';
                html += '        </a>';
                html += '      </div>';
                html += '    </div>';
                html += '  </div>';
                html += '  <div class="insight-files border-t border-border">';
                html += '    <div class="insight-files-content p-4 bg-muted/30">';

                // For duplicates, group files by duplicate set
                if (category === 'duplicates') {
                    html += renderDuplicateGroups(group.items);
                } else {
                    html += '      <div class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-2">Affected Files</div>';
                    html += '      <ul class="space-y-1.5">';

                    // Collect all unique files from all items in this category
                    const allFiles = new Set();
                    group.items.forEach(item => {
                        if (item.files) {
                            item.files.forEach(file => allFiles.add(file));
                        }
                    });

                    Array.from(allFiles).forEach(file => {
                        const truncated = truncatePath(file, 80);
                        html += '<li title="' + file + '" class="px-3 py-2 text-xs font-mono text-muted-foreground bg-background/50 border border-border rounded transition-all hover:bg-background hover:border-primary/50 hover:text-foreground">' + truncated + '</li>';
                    });

                    html += '      </ul>';
                }

                html += '    </div>';
                html += '  </div>';
                html += '</div>';
            });

            container.innerHTML = html;
        }

        // Toggle insight card expansion with dynamic height
        function toggleInsight(index) {
            const card = document.getElementById('insight-' + index);
            const filesContainer = card.querySelector('.insight-files');
            const content = filesContainer.querySelector('.insight-files-content');

            if (card.classList.contains('expanded')) {
                // Collapse
                filesContainer.style.maxHeight = '0';
                card.classList.remove('expanded');
            } else {
                // Expand - calculate actual content height
                const height = content.scrollHeight;
                filesContainer.style.maxHeight = height + 'px';
                card.classList.add('expanded');
            }
        }

        // Render duplicate files grouped by duplicate set
        function renderDuplicateGroups(items) {
            let html = '';

            // Sort items by impact (wasted size) descending
            const sortedItems = [...items].sort((a, b) => b.impact - a.impact);

            sortedItems.forEach((item, idx) => {
                if (!item.files || item.files.length === 0) return;

                // Extract filename from first file path for the group header
                const firstFile = item.files[0];
                const filename = firstFile.split('/').pop();
                const copyCount = item.files.length;
                const wastedSize = formatBytes(item.impact);

                html += '<div class="mb-4 pb-3 border-b border-border last:mb-0 last:pb-0 last:border-0">';
                html += '  <div class="flex justify-between items-center mb-2 p-2 bg-background/50 rounded flex-wrap gap-2">';
                html += '    <span class="font-semibold text-xs font-mono truncate max-w-[60%%]" title="' + filename + '">' + filename + '</span>';
                html += '    <span class="text-xs text-muted-foreground bg-muted px-2 py-0.5 rounded whitespace-nowrap">' + copyCount + ' copies · ' + wastedSize + '</span>';
                html += '  </div>';
                html += '  <ul class="space-y-1.5">';

                item.files.forEach(file => {
                    const truncated = truncatePath(file, 80);
                    html += '<li title="' + file + '" class="px-3 py-2 text-xs font-mono text-muted-foreground bg-background/50 border border-border rounded transition-all hover:bg-background hover:border-primary/50 hover:text-foreground">' + truncated + '</li>';
                });

                html += '  </ul>';
                html += '</div>';
            });

            return html;
        }

        // Initialize visualizations
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize theme first
            initTheme();

            // Store original data for search filtering (deep copy)
            if (reportData.fileTree) {
                originalFileTree = JSON.parse(JSON.stringify(reportData.fileTree));
            }
            if (reportData.categories) {
                originalCategories = JSON.parse(JSON.stringify(reportData.categories));
            }
            if (reportData.extensions) {
                originalExtensions = JSON.parse(JSON.stringify(reportData.extensions));
            }

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

        // Keyboard shortcut: Press 'D' to toggle dark mode
        document.addEventListener('keydown', function(event) {
            // Check if the key is 'd' or 'D' and not in an input field
            if ((event.key === 'd' || event.key === 'D') &&
                !['INPUT', 'TEXTAREA'].includes(event.target.tagName)) {
                event.preventDefault();
                toggleTheme();
            }
        });

        // Parse search query to detect special syntax
        function parseSearchQuery(query) {
            if (!query) {
                return { mode: 'empty', query: '' };
            }

            // Check for backtick syntax: ` + "`" + `moduleName` + "`" + `
            const backtickMatch = query.match(/` + "`" + `([^` + "`" + `]+)` + "`" + `/);
            if (backtickMatch) {
                return { mode: 'backtick', query: backtickMatch[1].toLowerCase() };
            }

            // Check for path-specific syntax: contains /
            if (query.includes('/')) {
                return { mode: 'path', query: query.toLowerCase() };
            }

            // Default: basic search
            return { mode: 'basic', query: query.toLowerCase() };
        }

        // Deep copy a tree node
        function deepCopyNode(node) {
            const copy = {
                name: node.name,
                value: node.value
            };
            if (node.path) copy.path = node.path;
            if (node.fileType) copy.fileType = node.fileType;
            if (node.itemStyle) copy.itemStyle = JSON.parse(JSON.stringify(node.itemStyle));
            if (node.isDuplicate) copy.isDuplicate = node.isDuplicate;
            if (node.children && node.children.length > 0) {
                copy.children = node.children.map(deepCopyNode);
            }
            return copy;
        }

        // Filter tree based on search query
        function filterTreeByQuery(tree, searchMode, query) {
            if (!tree || searchMode === 'empty') {
                return tree;
            }

            // Backtick mode: find exact node match by name, return it with all children
            if (searchMode === 'backtick') {
                function findNodeByName(node, targetName) {
                    if (node.name.toLowerCase() === targetName) {
                        return deepCopyNode(node);
                    }
                    if (node.children) {
                        for (const child of node.children) {
                            const found = findNodeByName(child, targetName);
                            if (found) return found;
                        }
                    }
                    return null;
                }

                const foundNode = findNodeByName(tree, query);
                if (foundNode) {
                    // Wrap in a root node to maintain structure
                    return {
                        name: tree.name,
                        value: foundNode.value,
                        children: [foundNode]
                    };
                }
                return null;
            }

            // Path-specific mode: only include nodes whose path starts with the query
            if (searchMode === 'path') {
                function filterByPath(node) {
                    const nodePath = (node.path || '').toLowerCase();
                    const nodeName = node.name.toLowerCase();

                    // Check if this node's path matches
                    const pathMatches = nodePath.includes(query) || nodeName.includes(query);

                    if (!node.children || node.children.length === 0) {
                        // Leaf node: include if path matches
                        return pathMatches ? deepCopyNode(node) : null;
                    }

                    // Parent node: recursively filter children
                    const filteredChildren = node.children
                        .map(filterByPath)
                        .filter(child => child !== null);

                    if (filteredChildren.length > 0) {
                        const copy = deepCopyNode(node);
                        copy.children = filteredChildren;
                        // Recalculate value based on filtered children
                        copy.value = filteredChildren.reduce((sum, child) => sum + child.value, 0);
                        return copy;
                    }

                    return null;
                }

                const filtered = filterByPath(tree);
                return filtered;
            }

            // Basic mode: match against name, path, extension, fileType
            if (searchMode === 'basic') {
                function filterBasic(node) {
                    const nodeName = node.name.toLowerCase();
                    const nodePath = (node.path || '').toLowerCase();
                    const nodeType = (node.fileType || '').toLowerCase();

                    // Check if this node matches
                    const matches = nodeName.includes(query) ||
                                  nodePath.includes(query) ||
                                  nodeType.includes(query);

                    if (!node.children || node.children.length === 0) {
                        // Leaf node: include if matches
                        return matches ? deepCopyNode(node) : null;
                    }

                    // Parent node: recursively filter children
                    const filteredChildren = node.children
                        .map(filterBasic)
                        .filter(child => child !== null);

                    if (filteredChildren.length > 0 || matches) {
                        const copy = deepCopyNode(node);
                        if (filteredChildren.length > 0) {
                            copy.children = filteredChildren;
                            // Recalculate value based on filtered children
                            copy.value = filteredChildren.reduce((sum, child) => sum + child.value, 0);
                        }
                        return copy;
                    }

                    return null;
                }

                const filtered = filterBasic(tree);
                return filtered;
            }

            return tree;
        }

        // Calculate categories and extensions from filtered tree
        function calculateStatsFromTree(tree) {
            const categoryMap = {};
            const extensionMap = {};

            function traverse(node) {
                // Count file types for categories
                if (node.fileType && node.value) {
                    const type = node.fileType;
                    categoryMap[type] = (categoryMap[type] || 0) + node.value;
                }

                // Count extensions (only for leaf nodes)
                if ((!node.children || node.children.length === 0) && node.name && node.value) {
                    const lastDot = node.name.lastIndexOf('.');
                    if (lastDot > 0) {
                        const ext = node.name.substring(lastDot);
                        extensionMap[ext] = (extensionMap[ext] || 0) + node.value;
                    } else {
                        // Files without extension
                        extensionMap['(no ext)'] = (extensionMap['(no ext)'] || 0) + node.value;
                    }
                }

                if (node.children) {
                    node.children.forEach(traverse);
                }
            }

            traverse(tree);

            // Convert maps to sorted arrays
            const categories = Object.entries(categoryMap)
                .map(([name, value]) => ({ name, value }))
                .sort((a, b) => b.value - a.value);

            const extensions = Object.entries(extensionMap)
                .map(([name, value]) => ({ name, value }))
                .sort((a, b) => b.value - a.value)
                .slice(0, 10); // Top 10 extensions

            return { categories, extensions };
        }

        // Update charts with filtered data
        function updateChartsWithFilteredData(filteredTree, categories, extensions) {
            if (treemapChart && filteredTree) {
                const treemapOption = getTreemapOption(filteredTree);
                treemapChart.setOption(treemapOption, true);
            }
            if (categoryChart && categories && categories.length > 0) {
                const categoryOption = getCategoryChartOption(categories);
                categoryChart.setOption(categoryOption, true);
            }
            if (extensionChart && extensions && extensions.length > 0) {
                const extensionOption = getExtensionChartOption(extensions);
                extensionChart.setOption(extensionOption, true);
            }
        }

        // Search functionality with debouncing
        let searchTimeout = null;
        document.getElementById('search-input').addEventListener('input', function(e) {
            const query = e.target.value.trim();

            // Clear previous timeout
            if (searchTimeout) {
                clearTimeout(searchTimeout);
            }

            // Debounce search (300ms)
            searchTimeout = setTimeout(function() {
                const parsed = parseSearchQuery(query);

                // Empty search: restore original data
                if (parsed.mode === 'empty') {
                    if (originalFileTree) {
                        updateChartsWithFilteredData(
                            JSON.parse(JSON.stringify(originalFileTree)),
                            JSON.parse(JSON.stringify(originalCategories)),
                            JSON.parse(JSON.stringify(originalExtensions))
                        );
                    }
                    return;
                }

                // Filter the tree
                const filteredTree = filterTreeByQuery(
                    JSON.parse(JSON.stringify(originalFileTree)),
                    parsed.mode,
                    parsed.query
                );

                // If no results, show empty state
                if (!filteredTree || (filteredTree.children && filteredTree.children.length === 0)) {
                    // Create an empty tree structure
                    const emptyTree = {
                        name: 'No results',
                        value: 0,
                        children: []
                    };
                    updateChartsWithFilteredData(emptyTree, [], []);
                    return;
                }

                // Recalculate stats from filtered tree
                const stats = calculateStatsFromTree(filteredTree);

                // Update all charts
                updateChartsWithFilteredData(filteredTree, stats.categories, stats.extensions);
            }, 300);
        });
    </script>
</body>
</html>
`
