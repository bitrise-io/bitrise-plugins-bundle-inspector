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
            --popover: 0 0% 100%;
            --popover-foreground: 222.2 84% 4.9%;

            /* Bitrise Brand Colors for Category Headers */
            --color-header-0: #DEB4FF;  /* Base lavender purple */
            --color-header-1: #B190CC;  /* Darker 20% */
            --color-header-2: #856C99;  /* Darker 40% */
            --color-header-3: #584866;  /* Darker 60% */
            --color-header-4: #2C2433;  /* Darkest 80% */

            /* Content/File Type Colors - Emerge Tools Aligned */
            --color-framework: #5B7FDB;          /* Blue (was #9247C2) */
            --color-library: #0dd3c5;            /* Cyan (unchanged) */
            --color-native: #ff9500;             /* Orange (unchanged) */
            --color-image: #30d158;              /* Green (was #ffd60a yellow) */
            --color-asset-catalog: #59886b;      /* Forest Green (was #30d158) */
            --color-resource: #64d2ff;           /* Light Blue (unchanged) */
            --color-ui: #bf5af2;                 /* Magenta (unchanged) */
            --color-dex: #ac8e68;                /* Brown (unchanged) */
            --color-font: #6a097d;               /* Purple (NEW - fixes bug) */
            --color-video: #0e49b5;              /* Dark Blue (NEW) */
            --color-audio: #ff6b35;              /* Coral (NEW) */
            --color-mlmodel: #583D72;            /* Deep Purple (NEW) */
            --color-localization: #ffa45b;       /* Light Orange (NEW) */
            --color-duplicate: #ff453a;          /* Red (unchanged) */
            --color-other: #98989d;              /* Gray (unchanged) */
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
            --popover: 222.2 84% 4.9%;
            --popover-foreground: 210 40% 98%;
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

        /* Screen reader only utility */
        .sr-only {
            position: absolute;
            width: 1px;
            height: 1px;
            padding: 0;
            margin: -1px;
            overflow: hidden;
            clip: rect(0, 0, 0, 0);
            white-space: nowrap;
            border-width: 0;
        }

        /* shadcn/ui style tooltip */
        .tooltip-trigger {
            position: relative;
            display: inline-flex;
        }

        .tooltip-content {
            position: absolute;
            z-index: 50;
            bottom: calc(100% + 8px);
            left: 50%;
            transform: translateX(-50%);
            padding: 0.5rem 0.75rem;
            background: hsl(var(--popover));
            color: hsl(var(--popover-foreground));
            border: 1px solid hsl(var(--border));
            border-radius: calc(var(--radius) - 2px);
            box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
            font-size: 0.875rem;
            line-height: 1.25rem;
            white-space: nowrap;
            pointer-events: none;
            opacity: 0;
            transition: opacity 150ms ease-in-out;
            text-transform: none;
        }

        .dark .tooltip-content {
            box-shadow: 0 4px 6px -1px rgb(0 0 0 / 0.3), 0 2px 4px -2px rgb(0 0 0 / 0.3);
        }

        .tooltip-trigger:hover .tooltip-content,
        .tooltip-trigger:focus .tooltip-content {
            opacity: 1;
        }

        /* Tooltip arrow */
        .tooltip-content::after {
            content: '';
            position: absolute;
            top: 100%;
            left: 50%;
            transform: translateX(-50%);
            border: 4px solid transparent;
            border-top-color: hsl(var(--border));
        }

        .tooltip-content::before {
            content: '';
            position: absolute;
            top: 100%;
            left: 50%;
            transform: translateX(-50%);
            border: 3px solid transparent;
            border-top-color: hsl(var(--popover));
            z-index: 1;
        }

        /* Legend item styles */
        .legend-item {
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .legend-color {
            width: 1rem;
            height: 1rem;
            border-radius: 0.25rem;
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
                        <span class="font-semibold text-lg leading-none tracking-tight">Bundle Inspector</span>
                        <span class="text-xs text-muted-foreground leading-none">Size Analysis Report</span>
                    </div>
                </div>

                <!-- Dark Mode Toggle with Keyboard Shortcut -->
                <button data-action="toggle-theme"
                        class="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 hover:bg-accent hover:text-accent-foreground h-9 w-9 relative group"
                        aria-label="Toggle Mode (D)"
                        title="Toggle Mode ` + "`" + `D` + "`" + `">
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
        <div class="rounded-lg border bg-card text-card-foreground shadow-sm">
            <!-- Header Section -->
            <div class="bg-gradient-to-r from-primary/5 to-primary/10 px-6 py-5 border-b rounded-t-lg">
                <div class="flex items-center gap-4">
                    {{if .IconData}}
                    <div class="flex-shrink-0">
                        <img src="{{.IconData}}" alt="App Icon" class="w-16 h-16 rounded-xl shadow-md ring-2 ring-background" />
                    </div>
                    {{end}}
                    <div class="flex-grow min-w-0">
                        <h1 class="scroll-m-20 text-3xl font-bold tracking-tight truncate">{{if .AppName}}{{.AppName}}{{else}}{{.Title}}{{end}}</h1>
                        {{if .BundleID}}<p class="text-sm text-muted-foreground mt-1 font-mono truncate">{{.BundleID}}</p>{{end}}
                    </div>
                </div>
            </div>

            <!-- Stats Grid -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-px bg-border">
                <!-- Platform -->
                {{if .Platform}}
                <div class="bg-card px-5 py-4">
                    <div class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1 flex items-center gap-1.5">
                        <span>Platform</span>
                        <span class="tooltip-trigger inline-flex items-center justify-center w-3.5 h-3.5 rounded-full border border-muted-foreground/30 text-[10px] cursor-help hover:bg-muted transition-colors">
                            ?
                            <span class="tooltip-content">The operating system platform (iOS, Android)</span>
                        </span>
                    </div>
                    <div class="text-lg font-semibold">{{.Platform}}</div>
                </div>
                {{end}}

                <!-- Version -->
                {{if .Version}}
                <div class="bg-card px-5 py-4">
                    <div class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1 flex items-center gap-1.5">
                        <span>Version</span>
                        <span class="tooltip-trigger inline-flex items-center justify-center w-3.5 h-3.5 rounded-full border border-muted-foreground/30 text-[10px] cursor-help hover:bg-muted transition-colors">
                            ?
                            <span class="tooltip-content">The application version number</span>
                        </span>
                    </div>
                    <div class="text-lg font-semibold">{{.Version}}</div>
                </div>
                {{end}}

                <!-- Type -->
                {{if .ArtifactType}}
                <div class="bg-card px-5 py-4">
                    <div class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1 flex items-center gap-1.5">
                        <span>Type</span>
                        <span class="tooltip-trigger inline-flex items-center justify-center w-3.5 h-3.5 rounded-full border border-muted-foreground/30 text-[10px] cursor-help hover:bg-muted transition-colors">
                            ?
                            <span class="tooltip-content">IPA (iOS App Store), APK (Android Package), or AAB (Android App Bundle)</span>
                        </span>
                    </div>
                    <div class="text-lg font-semibold uppercase">{{.ArtifactType}}</div>
                </div>
                {{end}}

                <!-- Potential Savings (Highlighted) -->
                <div class="bg-card px-5 py-4">
                    <div class="text-xs font-medium text-muted-foreground uppercase tracking-wide mb-1 flex items-center gap-1.5">
                        <span>Savings</span>
                        <span class="tooltip-trigger inline-flex items-center justify-center w-3.5 h-3.5 rounded-full border border-muted-foreground/30 text-[10px] cursor-help hover:bg-muted transition-colors">
                            ?
                            <span class="tooltip-content">Potential size reduction from optimizations (duplicates, compression)</span>
                        </span>
                    </div>
                    <div class="text-lg font-bold text-green-600 dark:text-green-500">{{.TotalSavings}}</div>
                </div>
            </div>

            <!-- Size Metrics -->
            <div class="bg-card px-6 py-4 border-t">
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div class="flex items-center justify-between">
                        <span class="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
                            <span>Download Size</span>
                            <span class="tooltip-trigger inline-flex items-center justify-center w-3.5 h-3.5 rounded-full border border-muted-foreground/30 text-[10px] cursor-help hover:bg-muted transition-colors">
                                ?
                                <span class="tooltip-content">Size users download from the app store (compressed)</span>
                            </span>
                        </span>
                        <span class="text-xl font-semibold">{{.TotalSize}}</span>
                    </div>
                    <div class="flex items-center justify-between">
                        <span class="text-sm font-medium text-muted-foreground flex items-center gap-1.5">
                            <span>Install Size</span>
                            <span class="tooltip-trigger inline-flex items-center justify-center w-3.5 h-3.5 rounded-full border border-muted-foreground/30 text-[10px] cursor-help hover:bg-muted transition-colors">
                                ?
                                <span class="tooltip-content">Size on device after installation (uncompressed)</span>
                            </span>
                        </span>
                        <span class="text-xl font-semibold">{{.UncompressedSize}}</span>
                    </div>
                </div>
            </div>
        </div>

            <!-- Tabs -->
            <div class="space-y-4">
            <div class="inline-flex h-10 items-center justify-center rounded-md bg-muted p-1 text-muted-foreground">
                <button class="tab-button active bg-background text-foreground shadow-sm inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 gap-2"
                        data-action="switch-tab" data-tab="app-analyzer" title="App Analyzer ` + "`" + `A` + "`" + `">App Analyzer <kbd class="hidden sm:inline-flex h-5 select-none items-center rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">A</kbd></button>
                <button class="tab-button inline-flex items-center justify-center whitespace-nowrap rounded-sm px-3 py-1.5 text-sm font-medium ring-offset-background transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 gap-2"
                        data-action="switch-tab" data-tab="category" title="Category ` + "`" + `C` + "`" + `">Category <kbd class="hidden sm:inline-flex h-5 select-none items-center rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">C</kbd></button>
            </div>

            <section id="app-analyzer-panel" class="tab-panel active" aria-labelledby="treemap-heading">
                <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6 hover:-translate-y-0.5 transition-transform duration-300">
                    <h2 id="treemap-heading" class="scroll-m-20 text-2xl font-semibold tracking-tight">Bundle Treemap</h2>
                    <p class="text-sm text-muted-foreground leading-relaxed mt-1.5 mb-4">Click to drill down into folders. Use mouse wheel to zoom. Use breadcrumb to navigate back.</p>
                    <div class="mb-4">
                        <label for="search-input" class="sr-only">Search files</label>
                        <div class="relative">
                            <svg class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"></path>
                            </svg>
                            <input type="text" id="search-input"
                                   class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 pl-10 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                                   placeholder="Search files (e.g., .png, Frameworks/, &#96;Assets.car&#96;)"
                                   aria-describedby="search-hint">
                            <span id="search-hint" class="sr-only">Use backticks for exact module match, forward slash for path search</span>
                        </div>
                    </div>
                    <div id="treemap" role="img" aria-label="Bundle size treemap visualization"></div>
                    <div id="legend-container" class="flex flex-wrap gap-4 mt-4 pt-4 border-t border-border" role="list" aria-label="File type legend"></div>
                </div>
            </section>

            <section id="category-panel" class="tab-panel" aria-labelledby="category-heading">
                <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6 hover:-translate-y-0.5 transition-transform duration-300">
                        <h2 id="category-heading" class="scroll-m-20 text-xl font-semibold tracking-tight mb-6">Category Breakdown</h2>
                        <div id="category-chart" class="chart" role="img" aria-label="Category breakdown pie chart"></div>
                    </div>
                    <div class="rounded-lg border bg-card text-card-foreground shadow-sm p-6 hover:-translate-y-0.5 transition-transform duration-300">
                        <h2 class="scroll-m-20 text-xl font-semibold tracking-tight mb-6">Top Extensions</h2>
                        <div id="extension-chart" class="chart" role="img" aria-label="Top file extensions bar chart"></div>
                    </div>
                </div>
            </section>
        </div>

        <!-- Insights Section -->
        <section class="rounded-lg border bg-card text-card-foreground shadow-sm p-6" id="insights-section" aria-labelledby="insights-heading">
            <div class="flex items-center justify-between mb-6">
                <h2 id="insights-heading" class="scroll-m-20 text-2xl font-semibold tracking-tight flex items-center gap-3">
                    <svg xmlns="http://www.w3.org/2000/svg" width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="text-primary" aria-hidden="true"><path d="M9 18h6"/><path d="M10 22h4"/><path d="M12 2a7 7 0 0 0-4 12.7V17a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1v-2.3A7 7 0 0 0 12 2z"/></svg>
                    Insights & Optimization Opportunities
                </h2>
                <button data-action="toggle-all-insights"
                        class="inline-flex items-center justify-center gap-2 rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 hover:bg-accent hover:text-accent-foreground h-9 px-3"
                        title="Expand/Collapse All ` + "`" + `E` + "`" + `"
                        aria-label="Expand or collapse all insights (E)">
                    <svg id="expand-all-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m7 20 5-5 5 5"/><path d="m7 4 5 5 5-5"/></svg>
                    <span class="hidden sm:inline">Expand All</span>
                    <kbd class="hidden sm:inline-flex h-5 select-none items-center rounded border bg-muted px-1.5 font-mono text-[10px] font-medium text-muted-foreground">E</kbd>
                </button>
            </div>
            <div id="insights-list" class="space-y-4" role="list"></div>
        </section>

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

        // ============================================================
        // SafeHTML - Utility for safe DOM manipulation (XSS prevention)
        // ============================================================
        const SafeHTML = {
            // Escape HTML entities to prevent XSS
            escapeText(str) {
                if (str == null) return '';
                const div = document.createElement('div');
                div.textContent = String(str);
                return div.innerHTML;
            },

            // Create an element safely with attributes and children
            createElement(tag, attrs = {}, children = []) {
                const el = document.createElement(tag);

                for (const [key, value] of Object.entries(attrs)) {
                    if (key === 'className') {
                        el.className = value;
                    } else if (key === 'innerHTML') {
                        // Skip innerHTML - use children or textContent instead
                        console.warn('SafeHTML: Use children or textContent instead of innerHTML');
                    } else if (key === 'textContent') {
                        el.textContent = value;
                    } else if (key.startsWith('data-')) {
                        el.setAttribute(key, value);
                    } else if (key === 'style' && typeof value === 'object') {
                        Object.assign(el.style, value);
                    } else if (typeof value === 'boolean') {
                        if (value) el.setAttribute(key, '');
                    } else {
                        el.setAttribute(key, value);
                    }
                }

                for (const child of children) {
                    if (typeof child === 'string') {
                        el.appendChild(document.createTextNode(child));
                    } else if (child instanceof Node) {
                        el.appendChild(child);
                    }
                }

                return el;
            },

            // Create a text node
            text(str) {
                return document.createTextNode(str == null ? '' : String(str));
            }
        };

        // Safe element getter with null check
        function safeGetElement(id) {
            const el = document.getElementById(id);
            if (!el) {
                console.warn('Element not found: ' + id);
            }
            return el;
        }

        // URL validation for external links
        const ALLOWED_URL_DOMAINS = [
            'devcenter.bitrise.io',
            'bitrise.io',
            'github.com',
            'developer.apple.com',
            'developer.android.com'
        ];

        function isValidLearnMoreURL(url) {
            if (!url || typeof url !== 'string') return false;
            try {
                const parsed = new URL(url);
                // Only allow https
                if (parsed.protocol !== 'https:') return false;
                // Check against allowed domains
                return ALLOWED_URL_DOMAINS.some(domain =>
                    parsed.hostname === domain || parsed.hostname.endsWith('.' + domain)
                );
            } catch {
                return false;
            }
        }

        // ============================================================
        // Application State (Consolidated)
        // ============================================================
        const AppState = {
            theme: 'light',
            charts: {
                treemap: null,
                category: null,
                extension: null
            },
            originalData: {
                fileTree: null,
                categories: null,
                extensions: null
            },
            duplicatePaths: new Set(reportData.duplicates || []),
            dataTableStates: {},
            searchCache: new Map(), // Cache search results for performance
            maxCacheSize: 20,

            isDark() {
                return this.theme === 'dark';
            },

            // Get cached search result or null
            getCachedSearch(query) {
                return this.searchCache.get(query) || null;
            },

            // Cache a search result (LRU eviction)
            setCachedSearch(query, result) {
                if (this.searchCache.size >= this.maxCacheSize) {
                    // Remove oldest entry (first key)
                    const firstKey = this.searchCache.keys().next().value;
                    this.searchCache.delete(firstKey);
                }
                this.searchCache.set(query, result);
            },

            // Clear search cache (call when data changes)
            clearSearchCache() {
                this.searchCache.clear();
            },

            setTheme(theme) {
                this.theme = theme;
                if (theme === 'dark') {
                    document.documentElement.classList.add('dark');
                } else {
                    document.documentElement.classList.remove('dark');
                }
                localStorage.setItem('theme', theme);
            },

            init() {
                this.theme = localStorage.getItem('theme') || 'light';
                if (this.theme === 'dark') {
                    document.documentElement.classList.add('dark');
                }
                // Store original data for search filtering (deep copy)
                if (reportData.fileTree) {
                    this.originalData.fileTree = JSON.parse(JSON.stringify(reportData.fileTree));
                }
                if (reportData.categories) {
                    this.originalData.categories = JSON.parse(JSON.stringify(reportData.categories));
                }
                if (reportData.extensions) {
                    this.originalData.extensions = JSON.parse(JSON.stringify(reportData.extensions));
                }
            }
        };

        // Legacy aliases for backward compatibility during transition
        let currentTheme = 'light';
        let treemapChart = null;
        let categoryChart = null;
        let extensionChart = null;
        let originalFileTree = null;
        let originalCategories = null;
        let originalExtensions = null;

        // ============================================================
        // TreeUtils - Tree traversal and manipulation utilities
        // ============================================================
        const TreeUtils = {
            // Deep copy a tree node
            deepCopy(node) {
                if (!node) return null;
                const copy = {
                    name: node.name,
                    value: node.value
                };
                if (node.path) copy.path = node.path;
                if (node.fileType) copy.fileType = node.fileType;
                if (node.itemStyle) copy.itemStyle = JSON.parse(JSON.stringify(node.itemStyle));
                if (node.isDuplicate) copy.isDuplicate = node.isDuplicate;
                if (node.children && node.children.length > 0) {
                    copy.children = node.children.map(child => TreeUtils.deepCopy(child));
                }
                return copy;
            },

            // Traverse tree with visitor function
            traverse(node, visitor, depth = 0) {
                if (!node) return;
                visitor(node, depth);
                if (node.children) {
                    node.children.forEach(child => TreeUtils.traverse(child, visitor, depth + 1));
                }
            },

            // Collect nodes matching a predicate
            collect(node, predicate) {
                const results = [];
                TreeUtils.traverse(node, (n) => {
                    if (predicate(n)) results.push(n);
                });
                return results;
            },

            // Filter tree, keeping only matching nodes and their parents
            filter(node, predicate) {
                if (!node) return null;

                const matches = predicate(node);
                const isLeaf = !node.children || node.children.length === 0;

                if (isLeaf) {
                    return matches ? TreeUtils.deepCopy(node) : null;
                }

                const filteredChildren = node.children
                    .map(child => TreeUtils.filter(child, predicate))
                    .filter(child => child !== null);

                if (filteredChildren.length > 0 || matches) {
                    const copy = TreeUtils.deepCopy(node);
                    if (filteredChildren.length > 0) {
                        copy.children = filteredChildren;
                        copy.value = filteredChildren.reduce((sum, child) => sum + child.value, 0);
                    }
                    return copy;
                }

                return null;
            },

            // Find a node by name (exact match)
            findByName(node, targetName) {
                if (!node) return null;
                if (node.name.toLowerCase() === targetName.toLowerCase()) {
                    return TreeUtils.deepCopy(node);
                }
                if (node.children) {
                    for (const child of node.children) {
                        const found = TreeUtils.findByName(child, targetName);
                        if (found) return found;
                    }
                }
                return null;
            },

            // Aggregate values using a custom aggregator function
            aggregate(node, aggregator, initialValue = 0) {
                let result = initialValue;
                TreeUtils.traverse(node, (n) => {
                    result = aggregator(result, n);
                });
                return result;
            }
        };

        // ============================================================
        // ChartFactory - Unified chart management
        // ============================================================
        const ChartFactory = {
            instances: new Map(),
            resizeHandler: null,

            // Create or get a chart instance
            create(containerId, optionFn, data) {
                const container = safeGetElement(containerId);
                if (!container) return null;

                // Dispose existing instance
                if (this.instances.has(containerId)) {
                    this.instances.get(containerId).dispose();
                }

                const chartOptions = containerId === 'treemap'
                    ? { renderer: 'canvas', useDirtyRect: true }
                    : {};

                const chart = echarts.init(container, null, chartOptions);
                const option = optionFn(data);
                chart.setOption(option);

                this.instances.set(containerId, chart);
                this.ensureResizeHandler();

                return chart;
            },

            // Update an existing chart
            update(containerId, optionFn, data) {
                const chart = this.instances.get(containerId);
                if (chart) {
                    const option = optionFn(data);
                    chart.setOption(option, true);
                }
            },

            // Resize a specific chart
            resize(containerId) {
                const chart = this.instances.get(containerId);
                if (chart) {
                    chart.resize();
                }
            },

            // Resize all charts
            resizeAll() {
                this.instances.forEach(chart => chart.resize());
            },

            // Ensure single resize handler is registered
            ensureResizeHandler() {
                if (this.resizeHandler) return;

                let resizeTimeout;
                this.resizeHandler = () => {
                    clearTimeout(resizeTimeout);
                    resizeTimeout = setTimeout(() => this.resizeAll(), 100);
                };
                window.addEventListener('resize', this.resizeHandler);
            },

            // Dispose all charts
            disposeAll() {
                this.instances.forEach(chart => chart.dispose());
                this.instances.clear();
                if (this.resizeHandler) {
                    window.removeEventListener('resize', this.resizeHandler);
                    this.resizeHandler = null;
                }
            }
        };

        // ============================================================
        // Shared Chart Configuration
        // ============================================================
        function getBaseTooltipConfig() {
            const isDark = AppState.isDark();
            const themeColors = getThemeColors();
            return {
                backgroundColor: isDark ? '#2a2a2a' : '#fff',
                borderColor: isDark ? '#3a3a3a' : '#e5e5e7',
                textStyle: {
                    color: themeColors.textColor
                }
            };
        }

        // Generate treemap levels configuration
        function generateTreemapLevels(isDark) {
            const levels = [
                { itemStyle: { borderWidth: 0, gapWidth: 4 }, upperLabel: { show: false } }
            ];

            const levelConfigs = [
                { gapWidth: 2, borderWidth: 2, borderColor: isDark ? '#444' : '#ddd', height: 28, fontSize: 13, opacity: isDark ? 0.6 : 0.7 },
                { gapWidth: 2, borderWidth: 1, borderColor: isDark ? '#555' : '#eee', height: 24, fontSize: 12, opacity: isDark ? 0.5 : 0.6 },
                { gapWidth: 1, borderWidth: 1, borderColor: isDark ? '#555' : '#eee', height: 22, fontSize: 11, opacity: isDark ? 0.4 : 0.5 },
                { gapWidth: 1, borderWidth: 1, borderColor: undefined, height: 20, fontSize: 10, opacity: isDark ? 0.35 : 0.45 },
                { gapWidth: 1, borderWidth: 1, borderColor: undefined, height: 18, fontSize: 9, opacity: isDark ? 0.3 : 0.4 }
            ];

            levelConfigs.forEach(cfg => {
                const level = {
                    itemStyle: {
                        gapWidth: cfg.gapWidth,
                        borderWidth: cfg.borderWidth
                    },
                    upperLabel: {
                        show: true,
                        height: cfg.height,
                        formatter: function(params) {
                            return '{bg|' + params.name + '}';
                        },
                        rich: {
                            bg: {
                                backgroundColor: 'rgba(0,0,0,' + cfg.opacity + ')',
                                color: '#fff',
                                fontWeight: cfg.fontSize >= 12 ? 'bold' : 'normal',
                                fontSize: cfg.fontSize,
                                padding: cfg.fontSize >= 12 ? [4, 8] : (cfg.fontSize >= 10 ? [2, 5] : [1, 3]),
                                borderRadius: cfg.fontSize >= 11 ? 3 : 2
                            }
                        }
                    }
                };
                if (cfg.borderColor) {
                    level.itemStyle.borderColor = cfg.borderColor;
                }
                levels.push(level);
            });

            return levels;
        }

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

            const selectedPanel = safeGetElement(tabName + '-panel');
            if (selectedPanel) selectedPanel.classList.add('active');

            // Resize charts when switching tabs
            if (tabName === 'category') {
                setTimeout(() => {
                    ChartFactory.resize('category-chart');
                    ChartFactory.resize('extension-chart');
                }, 100);
            } else if (tabName === 'app-analyzer') {
                setTimeout(() => {
                    ChartFactory.resize('treemap');
                }, 100);
            }
        }

        // Initialize theme from localStorage (legacy wrapper)
        function initTheme() {
            // Now handled by AppState.init()
        }

        // Toggle theme
        function toggleTheme() {
            const newTheme = AppState.theme === 'light' ? 'dark' : 'light';
            AppState.setTheme(newTheme);
            currentTheme = newTheme; // Sync legacy variable
            updateThemeButton();
            updateChartsTheme();
        }

        // Update theme button icon (rotate for visual feedback)
        function updateThemeButton() {
            const buttons = document.querySelectorAll('[data-action="toggle-theme"]');
            buttons.forEach(button => {
                const svg = button.querySelector('svg');
                if (svg) {
                    if (AppState.isDark()) {
                        svg.style.transform = 'rotate(180deg)';
                        button.setAttribute('aria-label', 'Toggle Mode (D)');
                        button.setAttribute('title', 'Toggle Mode ` + "`" + `D` + "`" + `');
                    } else {
                        svg.style.transform = 'rotate(0deg)';
                        button.setAttribute('aria-label', 'Toggle Mode (D)');
                        button.setAttribute('title', 'Toggle Mode ` + "`" + `D` + "`" + `');
                    }
                }
            });
        }

        // Update all charts with new theme
        function updateChartsTheme() {
            ChartFactory.update('treemap', getTreemapOption, reportData.fileTree);
            ChartFactory.update('category-chart', getCategoryChartOption, reportData.categories);
            ChartFactory.update('extension-chart', getExtensionChartOption, reportData.extensions);
        }

        // Get theme colors for ECharts
        function getThemeColors() {
            const isDark = AppState.isDark();
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
                'font': getCSSVariable('--color-font'),
                'video': getCSSVariable('--color-video'),
                'audio': getCSSVariable('--color-audio'),
                'mlmodel': getCSSVariable('--color-mlmodel'),
                'localization': getCSSVariable('--color-localization'),
                'other': getCSSVariable('--color-other'),
                'duplicate': getCSSVariable('--color-duplicate')
            };
        }

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

        // Calculate brightness of a hex color (0-255)
        // Uses the perceived brightness formula
        function getColorBrightness(hex) {
            // Remove # if present
            hex = hex.replace(/^#/, '');

            // Parse RGB
            const r = parseInt(hex.substring(0, 2), 16);
            const g = parseInt(hex.substring(2, 4), 16);
            const b = parseInt(hex.substring(4, 6), 16);

            // Calculate perceived brightness
            // Formula: (R * 299 + G * 587 + B * 114) / 1000
            return (r * 299 + g * 587 + b * 114) / 1000;
        }

        // Get text color (black or white) based on background brightness
        function getTextColorForBackground(bgColor) {
            const brightness = getColorBrightness(bgColor);
            // If brightness > 128, use dark text; otherwise use white text
            return brightness > 128 ? '#000' : '#fff';
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

        // Apply colors to tree nodes
        function applyColorsToTree(node, depth) {
            depth = depth || 0;
            const isParent = node.children && node.children.length > 0;

            // Determine the color for this node
            let nodeColor;

            if (isParent) {
                // Parent/group node - use Bitrise brand color based on depth
                const headerLevel = Math.min(depth, 4);
                nodeColor = getCSSVariable('--color-header-' + headerLevel);
            } else {
                // Leaf node - use file type color (no depth-based darkening)
                const colors = getFileTypeColors();
                if (node.path && AppState.duplicatePaths.has(node.path)) {
                    nodeColor = colors['duplicate'];
                    node.isDuplicate = true;
                } else {
                    const fileType = node.fileType || getDominantFileType(node);
                    nodeColor = getColorForFileType(fileType);
                }
            }

            // Apply color
            node.itemStyle = node.itemStyle || {};
            node.itemStyle.color = nodeColor;

            // Set text color based on background brightness
            const textColor = getTextColorForBackground(nodeColor);
            node.label = node.label || {};
            node.label.color = textColor;

            // For parent nodes, set a darker border to help define the header area
            if (isParent) {
                node.itemStyle.borderColor = darkenColor(nodeColor, 0.2);
            }

            if (node.children) {
                node.children.forEach(child => applyColorsToTree(child, depth + 1));
            }
        }

        // Get treemap option with theme support
        function getTreemapOption(data) {
            // Apply colors
            applyColorsToTree(data);

            const isDark = AppState.isDark();
            const tooltipConfig = getBaseTooltipConfig();
            const borderColor = isDark ? '#666' : '#fff';
            const emphasisBorder = isDark ? '#fff' : '#333';

            return {
                tooltip: {
                    ...tooltipConfig,
                    formatter: function(info) {
                        const value = info.value;
                        const name = SafeHTML.escapeText(info.name);
                        const path = SafeHTML.escapeText(info.data.path || info.name);
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
                            result += '<br/><span style="color: var(--color-duplicate); font-weight: bold;">⚠ Duplicate file</span>';
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
                    leafDepth: 5,
                    zoomToNodeRatio: 0.32 * 0.32,
                    scaleLimit: { min: 0.5, max: 20 },
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
                            textStyle: { color: '#fff', fontSize: 12 }
                        },
                        emphasis: {
                            itemStyle: {
                                color: isDark ? 'rgba(80,80,80,1)' : 'rgba(60,60,60,1)',
                                textStyle: { color: '#fff' }
                            }
                        }
                    },
                    label: {
                        show: true,
                        formatter: '{b}',
                        fontSize: 11,
                        overflow: 'truncate'
                        // Color is set per-node based on background brightness
                    },
                    itemStyle: {
                        borderColor: borderColor,
                        borderWidth: 2,
                        gapWidth: 2
                    },
                    emphasis: {
                        label: { show: true, fontSize: 12, fontWeight: 'bold' },
                        itemStyle: { borderColor: emphasisBorder, borderWidth: 3 }
                    },
                    visibleMin: 200,
                    childrenVisibleMin: 100,
                    levels: generateTreemapLevels(isDark),
                    data: [data]
                }]
            };
        }

        // Create treemap visualization
        function createTreemap(data) {
            const chart = ChartFactory.create('treemap', getTreemapOption, data);
            // Keep legacy reference for backward compatibility
            treemapChart = chart;
            AppState.charts.treemap = chart;
            return chart;
        }

        // Get category chart option with theme support
        function getCategoryChartOption(categories) {
            const isDark = AppState.isDark();
            const tooltipConfig = getBaseTooltipConfig();
            const themeColors = getThemeColors();

            return {
                tooltip: {
                    ...tooltipConfig,
                    trigger: 'item',
                    formatter: '{b}: {c} ({d}%%)'
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
                            return SafeHTML.escapeText(params.name) + '\n' + formatBytes(params.value);
                        },
                        fontSize: 11,
                        color: themeColors.textColor
                    },
                    emphasis: {
                        label: { show: true, fontSize: 13, fontWeight: 'bold' }
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
            const chart = ChartFactory.create('category-chart', getCategoryChartOption, categories);
            // Keep legacy reference for backward compatibility
            categoryChart = chart;
            AppState.charts.category = chart;
            return chart;
        }

        // Get extension chart option with theme support
        function getExtensionChartOption(extensions) {
            const isDark = AppState.isDark();
            const tooltipConfig = getBaseTooltipConfig();
            const themeColors = getThemeColors();

            return {
                tooltip: {
                    ...tooltipConfig,
                    trigger: 'axis',
                    axisPointer: { type: 'shadow' },
                    formatter: function(params) {
                        const data = params[0];
                        return SafeHTML.escapeText(data.name) + ': ' + formatBytes(data.value);
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
                            { offset: 0, color: getCSSVariable('--color-framework') || '#9247C2' },
                            { offset: 1, color: getCSSVariable('--color-library') || '#0dd3c5' }
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
            const chart = ChartFactory.create('extension-chart', getExtensionChartOption, extensions);
            // Keep legacy reference for backward compatibility
            extensionChart = chart;
            AppState.charts.extension = chart;
            return chart;
        }

        // SVG icons for categories
        const icons = {
            lightbulb: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 18h6"/><path d="M10 22h4"/><path d="M12 2a7 7 0 0 0-4 12.7V17a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1v-2.3A7 7 0 0 0 12 2z"/></svg>',
            wrench: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>',
            package: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M16.5 9.4l-9-5.19"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
            copy: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>',
            image: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><circle cx="8.5" cy="8.5" r="1.5"/><polyline points="21 15 16 10 5 21"/></svg>',
            camera: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"/><circle cx="12" cy="13" r="4"/></svg>',
            trash: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>',
            checkCircle: '<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>'
        };

        // Category metadata with icons and learn more links
        const categoryMetadata = {
            'strip-symbols': {
                icon: icons.wrench,
                title: 'Strip Binary Symbols',
                description: 'Remove debug symbols from binaries to reduce file size without affecting app functionality.',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/strip-debug-symbols.html'
            },
            'frameworks': {
                icon: icons.package,
                title: 'Unused Frameworks',
                description: 'Frameworks included in your bundle but not referenced or used by your application.',
                learnMore: 'https://devcenter.bitrise.io/en/builds/build-cache.html'
            },
            'duplicates': {
                icon: icons.copy,
                title: 'Duplicate Files',
                description: 'Identical files appearing multiple times in your bundle, wasting storage space.',
                learnMore: 'https://devcenter.bitrise.io/en/builds/build-cache.html'
            },
            'image-optimization': {
                icon: icons.image,
                title: 'Image Optimization',
                description: 'Images that can be compressed or converted to more efficient formats to reduce size.',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/optimizing-app-size.html'
            },
            'loose-images': {
                icon: icons.camera,
                title: 'Loose Images',
                description: 'Images stored as individual files instead of being compiled into asset catalogs.',
                learnMore: 'https://devcenter.bitrise.io/en/deploying/ios-deployment/optimizing-app-size.html'
            },
            'unnecessary-files': {
                icon: icons.trash,
                title: 'Unnecessary Files',
                description: 'Files that are not needed in production builds and can be safely removed.',
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

        // Render empty insights (no optimizations found)
        function renderEmptyInsights(container) {
            container.textContent = '';

            const wrapper = SafeHTML.createElement('div', {
                className: 'flex flex-col items-center justify-center py-10 gap-3'
            });

            // Success icon
            const iconSpan = SafeHTML.createElement('span', {
                className: 'text-success w-12 h-12'
            });
            iconSpan.innerHTML = icons.checkCircle.replace('width="24" height="24"', 'width="48" height="48"');
            wrapper.appendChild(iconSpan);

            // Success message
            wrapper.appendChild(SafeHTML.createElement('span', {
                className: 'text-lg font-semibold text-success',
                textContent: 'No optimization opportunities found!'
            }));

            // Subtitle
            wrapper.appendChild(SafeHTML.createElement('span', {
                className: 'text-sm text-muted-foreground',
                textContent: 'Your bundle is well optimized.'
            }));

            container.appendChild(wrapper);
        }

        // Create an insight card element using DOM APIs
        function createInsightCard(category, group, metadata, index, totalSize) {
            const savingsPercentage = totalSize > 0
                ? ((group.totalSavings / totalSize) * 100).toFixed(2)
                : '0.00';

            const card = SafeHTML.createElement('div', {
                className: 'insight-card rounded-lg border bg-card overflow-hidden transition-all duration-200 hover:shadow-md',
                id: 'insight-' + index,
                'data-action': 'toggle-insight',
                'data-index': String(index)
            });

            // Header section
            const header = SafeHTML.createElement('div', {
                className: 'flex items-start gap-3 p-4 cursor-pointer select-none'
            });

            // Icon container
            const iconDiv = SafeHTML.createElement('div', {
                className: 'flex-shrink-0 w-6 h-6 text-primary'
            });
            iconDiv.innerHTML = metadata.icon; // SVG icons are trusted internal data
            header.appendChild(iconDiv);

            // Content section
            const content = SafeHTML.createElement('div', { className: 'flex-1 min-w-0' });

            // Title row
            const titleRow = SafeHTML.createElement('div', {
                className: 'flex items-center justify-between gap-2 mb-1.5'
            });
            titleRow.appendChild(SafeHTML.createElement('h3', {
                className: 'text-sm font-semibold leading-none tracking-tight',
                textContent: metadata.title
            }));

            // Expand indicator
            const expandSvg = SafeHTML.createElement('svg', {
                className: 'expand-indicator w-4 h-4 text-muted-foreground flex-shrink-0 transition-transform duration-200',
                fill: 'none',
                stroke: 'currentColor',
                viewBox: '0 0 24 24'
            });
            expandSvg.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>';
            titleRow.appendChild(expandSvg);
            content.appendChild(titleRow);

            // Description
            content.appendChild(SafeHTML.createElement('p', {
                className: 'text-sm text-muted-foreground leading-normal mb-2.5',
                textContent: metadata.description || group.description
            }));

            // Stats row
            const statsRow = SafeHTML.createElement('div', {
                className: 'flex items-center gap-2 flex-wrap text-xs'
            });

            // Savings badge
            const savingsBadge = SafeHTML.createElement('span', {
                className: 'inline-flex items-center gap-1.5 font-semibold bg-success/10 text-success px-2 py-1 rounded-md'
            });
            savingsBadge.innerHTML = '<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/></svg>';
            savingsBadge.appendChild(SafeHTML.text(' ' + formatBytes(group.totalSavings) + ' (' + savingsPercentage + '%)'));
            statsRow.appendChild(savingsBadge);

            // Files badge
            const filesBadge = SafeHTML.createElement('span', {
                className: 'inline-flex items-center gap-1 text-muted-foreground px-2 py-1 bg-muted/50 rounded-md font-medium'
            });
            filesBadge.innerHTML = '<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z"/></svg>';
            filesBadge.appendChild(SafeHTML.text(' ' + group.totalFiles + ' files'));
            statsRow.appendChild(filesBadge);

            // Learn more link (with URL validation)
            if (isValidLearnMoreURL(metadata.learnMore)) {
                const learnMoreLink = SafeHTML.createElement('a', {
                    href: metadata.learnMore,
                    className: 'inline-flex items-center gap-0.5 text-primary hover:underline font-medium transition-colors',
                    target: '_blank',
                    rel: 'noopener noreferrer'
                });
                learnMoreLink.appendChild(SafeHTML.text('Learn more'));
                const arrowSvg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
                arrowSvg.setAttribute('class', 'w-3 h-3');
                arrowSvg.setAttribute('fill', 'none');
                arrowSvg.setAttribute('stroke', 'currentColor');
                arrowSvg.setAttribute('viewBox', '0 0 24 24');
                arrowSvg.innerHTML = '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>';
                learnMoreLink.appendChild(arrowSvg);
                statsRow.appendChild(learnMoreLink);
            }

            content.appendChild(statsRow);
            header.appendChild(content);
            card.appendChild(header);

            // Files section
            const filesSection = SafeHTML.createElement('div', {
                className: 'insight-files border-t border-border'
            });
            const filesContent = SafeHTML.createElement('div', {
                className: 'insight-files-content p-4 bg-muted/30'
            });

            // Render table content (still uses innerHTML for complex table rendering - to be improved in Phase 5)
            const tableId = 'table-' + category + '-' + index;
            if (category === 'duplicates') {
                filesContent.innerHTML = renderDuplicateGroups(group.items, tableId);
            } else {
                filesContent.innerHTML = renderFilesTable(group.items, tableId);
            }

            filesSection.appendChild(filesContent);
            card.appendChild(filesSection);

            return card;
        }

        // Render insights section
        function renderInsights(optimizations) {
            const container = safeGetElement('insights-list');
            if (!container) return;

            if (!optimizations || optimizations.length === 0) {
                renderEmptyInsights(container);
                return;
            }

            // Clear container
            container.textContent = '';

            const groups = groupByCategory(optimizations);
            const totalSize = reportData.fileTree ? reportData.fileTree.value : 0;

            // Render each category using DOM APIs
            Object.keys(groups).forEach((category, index) => {
                const group = groups[category];
                const metadata = categoryMetadata[category] || {
                    icon: icons.lightbulb,
                    title: category.replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase()),
                    learnMore: 'https://devcenter.bitrise.io'
                };

                const card = createInsightCard(category, group, metadata, index, totalSize);
                container.appendChild(card);
            });
        }

        // Toggle insight card expansion with dynamic height
        function toggleInsight(index) {
            const card = safeGetElement('insight-' + index);
            if (!card) return;
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

        // Track expand/collapse all state
        let allInsightsExpanded = false;

        // Toggle all insight cards
        function toggleAllInsights() {
            const cards = document.querySelectorAll('.insight-card');
            if (cards.length === 0) return;

            allInsightsExpanded = !allInsightsExpanded;

            cards.forEach(card => {
                const filesContainer = card.querySelector('.insight-files');
                const content = filesContainer.querySelector('.insight-files-content');

                if (allInsightsExpanded) {
                    // Expand
                    const height = content.scrollHeight;
                    filesContainer.style.maxHeight = height + 'px';
                    card.classList.add('expanded');
                } else {
                    // Collapse
                    filesContainer.style.maxHeight = '0';
                    card.classList.remove('expanded');
                }
            });

            // Update button text
            updateExpandAllButton();
        }

        // Update the expand/collapse all button text and icon
        function updateExpandAllButton() {
            const button = document.querySelector('[data-action="toggle-all-insights"]');
            if (!button) return;

            const textSpan = button.querySelector('span');
            if (textSpan) {
                textSpan.textContent = allInsightsExpanded ? 'Collapse All' : 'Expand All';
            }

            const icon = button.querySelector('svg');
            if (icon) {
                // Rotate icon when expanded
                icon.style.transform = allInsightsExpanded ? 'rotate(180deg)' : 'rotate(0deg)';
                icon.style.transition = 'transform 0.2s ease';
            }
        }

        // Switch to a specific tab programmatically
        function switchToTab(tabName) {
            const tabButton = document.querySelector('[data-action="switch-tab"][data-tab="' + tabName + '"]');
            if (!tabButton) return;

            // Update button styles
            const tabButtons = document.querySelectorAll('.tab-button');
            tabButtons.forEach(button => {
                button.classList.remove('active', 'bg-background', 'text-foreground', 'shadow-sm');
            });
            tabButton.classList.add('active', 'bg-background', 'text-foreground', 'shadow-sm');

            // Update panels
            const tabPanels = document.querySelectorAll('.tab-panel');
            tabPanels.forEach(panel => panel.classList.remove('active'));
            const selectedPanel = safeGetElement(tabName + '-panel');
            if (selectedPanel) selectedPanel.classList.add('active');

            // Resize charts
            if (tabName === 'category') {
                setTimeout(() => {
                    ChartFactory.resize('category-chart');
                    ChartFactory.resize('extension-chart');
                }, 100);
            } else if (tabName === 'app-analyzer') {
                setTimeout(() => {
                    ChartFactory.resize('treemap');
                }, 100);
            }
        }

        // Create a data table with sorting and pagination
        function createDataTable(tableId, data, columns, options = {}) {
            const pageSize = options.pageSize || 10;
            const state = {
                data: data,
                columns: columns,
                sortColumn: options.defaultSort || null,
                sortDirection: options.defaultSortDir || 'desc',
                currentPage: 0,
                pageSize: pageSize
            };
            AppState.dataTableStates[tableId] = state;

            // Sort data initially
            if (state.sortColumn) {
                sortDataTable(tableId, state.sortColumn, false);
            }

            return renderDataTable(tableId);
        }

        // Sort data table by column
        function sortDataTable(tableId, column, toggle = true) {
            const state = AppState.dataTableStates[tableId];
            if (!state) return;

            if (toggle && state.sortColumn === column) {
                state.sortDirection = state.sortDirection === 'asc' ? 'desc' : 'asc';
            } else if (toggle) {
                state.sortColumn = column;
                state.sortDirection = 'desc';
            }

            state.data.sort((a, b) => {
                let aVal = a[column];
                let bVal = b[column];

                // Handle string comparison
                if (typeof aVal === 'string') {
                    aVal = aVal.toLowerCase();
                    bVal = bVal.toLowerCase();
                }

                if (aVal < bVal) return state.sortDirection === 'asc' ? -1 : 1;
                if (aVal > bVal) return state.sortDirection === 'asc' ? 1 : -1;
                return 0;
            });

            state.currentPage = 0;
        }

        // Change page
        function changeDataTablePage(tableId, page) {
            const state = AppState.dataTableStates[tableId];
            if (!state) return;

            const maxPage = Math.ceil(state.data.length / state.pageSize) - 1;
            state.currentPage = Math.max(0, Math.min(page, maxPage));

            // Re-render the table
            const container = safeGetElement(tableId);
            if (container) {
                container.outerHTML = renderDataTable(tableId);
            }
        }

        // Handle sort click
        function handleDataTableSort(tableId, column) {
            sortDataTable(tableId, column, true);
            const container = safeGetElement(tableId);
            if (container) {
                container.outerHTML = renderDataTable(tableId);
            }
        }

        // Render data table HTML
        function renderDataTable(tableId) {
            const state = AppState.dataTableStates[tableId];
            if (!state) return '';

            const { data, columns, sortColumn, sortDirection, currentPage, pageSize } = state;
            const totalPages = Math.ceil(data.length / pageSize);
            const startIdx = currentPage * pageSize;
            const endIdx = Math.min(startIdx + pageSize, data.length);
            const pageData = data.slice(startIdx, endIdx);

            let html = '<div id="' + tableId + '" class="space-y-3">';

            // Table
            html += '<div class="rounded-md border overflow-hidden">';
            html += '<table class="w-full text-sm">';

            // Header
            html += '<thead class="bg-muted/50">';
            html += '<tr class="border-b border-border">';
            columns.forEach(col => {
                const isSorted = sortColumn === col.key;
                const sortIcon = isSorted
                    ? (sortDirection === 'asc'
                        ? '<svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 15l7-7 7 7"/></svg>'
                        : '<svg class="w-4 h-4 ml-1" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>')
                    : '<svg class="w-4 h-4 ml-1 opacity-0 group-hover:opacity-50" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"/></svg>';

                const alignClass = col.align === 'right' ? 'text-right justify-end' : 'text-left';
                const widthClass = col.width ? ' ' + col.width : '';

                if (col.sortable !== false) {
                    html += '<th class="h-10 px-4 align-middle font-medium text-muted-foreground' + widthClass + '">';
                    html += '<button onclick="handleDataTableSort(\'' + tableId + '\', \'' + col.key + '\')" class="group inline-flex items-center gap-1 hover:text-foreground transition-colors ' + alignClass + ' w-full">';
                    html += col.label + sortIcon;
                    html += '</button>';
                    html += '</th>';
                } else {
                    html += '<th class="h-10 px-4 align-middle font-medium text-muted-foreground ' + alignClass + widthClass + '">' + col.label + '</th>';
                }
            });
            html += '</tr>';
            html += '</thead>';

            // Body
            html += '<tbody class="divide-y divide-border">';
            pageData.forEach((row, idx) => {
                const rowClass = idx % 2 === 0 ? 'bg-background' : 'bg-muted/20';
                html += '<tr class="' + rowClass + ' hover:bg-muted/50 transition-colors">';
                columns.forEach(col => {
                    const alignClass = col.align === 'right' ? 'text-right' : 'text-left';
                    html += '<td class="p-4 align-top ' + alignClass + '">';
                    html += col.render ? col.render(row) : row[col.key];
                    html += '</td>';
                });
                html += '</tr>';
            });
            html += '</tbody>';
            html += '</table>';
            html += '</div>';

            // Pagination footer
            html += '<div class="flex items-center justify-between px-2">';
            html += '<div class="text-sm text-muted-foreground">';
            html += 'Showing ' + (startIdx + 1) + '-' + endIdx + ' of ' + data.length + ' files';
            html += '</div>';

            if (totalPages > 1) {
                html += '<div class="flex items-center gap-1">';

                // Previous button
                const prevDisabled = currentPage === 0;
                html += '<button onclick="changeDataTablePage(\'' + tableId + '\', ' + (currentPage - 1) + ')" ' + (prevDisabled ? 'disabled' : '') + ' class="inline-flex items-center justify-center rounded-md text-sm font-medium h-8 w-8 border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50">';
                html += '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>';
                html += '</button>';

                // Page numbers
                const maxVisiblePages = 5;
                let startPage = Math.max(0, currentPage - Math.floor(maxVisiblePages / 2));
                let endPage = Math.min(totalPages, startPage + maxVisiblePages);
                if (endPage - startPage < maxVisiblePages) {
                    startPage = Math.max(0, endPage - maxVisiblePages);
                }

                for (let i = startPage; i < endPage; i++) {
                    const isCurrentPage = i === currentPage;
                    const pageClass = isCurrentPage
                        ? 'bg-primary text-primary-foreground'
                        : 'border border-input bg-background hover:bg-accent hover:text-accent-foreground';
                    html += '<button onclick="changeDataTablePage(\'' + tableId + '\', ' + i + ')" class="inline-flex items-center justify-center rounded-md text-sm font-medium h-8 w-8 ' + pageClass + '">';
                    html += (i + 1);
                    html += '</button>';
                }

                // Next button
                const nextDisabled = currentPage >= totalPages - 1;
                html += '<button onclick="changeDataTablePage(\'' + tableId + '\', ' + (currentPage + 1) + ')" ' + (nextDisabled ? 'disabled' : '') + ' class="inline-flex items-center justify-center rounded-md text-sm font-medium h-8 w-8 border border-input bg-background hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50">';
                html += '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>';
                html += '</button>';

                html += '</div>';
            }

            html += '</div>';
            html += '</div>';

            return html;
        }

        // Render files as a data table
        function renderFilesTable(items, tableId) {
            // Collect all files with their estimated savings
            const fileMap = new Map();
            items.forEach(item => {
                if (item.files && item.files.length > 0) {
                    const perFileSavings = Math.floor(item.impact / item.files.length);
                    item.files.forEach(file => {
                        const existing = fileMap.get(file) || 0;
                        fileMap.set(file, existing + perFileSavings);
                    });
                }
            });

            // Convert to array
            const files = Array.from(fileMap.entries())
                .map(([path, savings]) => ({
                    path,
                    savings,
                    filename: path.split('/').pop()
                }));

            if (files.length === 0) {
                return '<p class="text-sm text-muted-foreground">No files to display.</p>';
            }

            const columns = [
                {
                    key: 'filename',
                    label: 'File',
                    sortable: true,
                    width: 'w-1/4',
                    render: (row) => {
                        return '<span class="text-sm font-medium truncate block" title="' + row.filename + '">' + row.filename + '</span>';
                    }
                },
                {
                    key: 'path',
                    label: 'Path',
                    sortable: false,
                    render: (row) => {
                        return '<div class="max-w-2xl"><div class="text-xs font-mono text-muted-foreground py-0.5 break-all">' + row.path + '</div></div>';
                    }
                },
                {
                    key: 'savings',
                    label: 'Savings',
                    align: 'right',
                    width: 'w-32',
                    sortable: true,
                    render: (row) => {
                        return '<span class="text-sm font-semibold text-success bg-success/10 px-2 py-1 rounded-md inline-block">' + formatBytes(row.savings) + '</span>';
                    }
                }
            ];

            return createDataTable(tableId, files, columns, {
                defaultSort: 'savings',
                defaultSortDir: 'desc',
                pageSize: 10
            });
        }

        // Render duplicate files grouped by duplicate set
        function renderDuplicateGroups(items, baseTableId) {
            // Sort items by impact (wasted size) descending
            const sortedItems = [...items].sort((a, b) => b.impact - a.impact);

            // Prepare data for single consolidated table
            const tableData = sortedItems.map(item => {
                if (!item.files || item.files.length === 0) return null;

                const firstFile = item.files[0];
                const filename = firstFile.split('/').pop();

                return {
                    file: filename,
                    locations: item.files.join(', '),
                    locationsArray: item.files, // For tooltip
                    count: item.files.length,
                    savings: item.impact,
                    savingsFormatted: formatBytes(item.impact)
                };
            }).filter(item => item !== null);

            const columns = [
                {
                    key: 'file',
                    label: 'File',
                    sortable: true,
                    width: 'w-1/4',
                    render: (row) => {
                        return '<div class="flex items-center gap-2">' +
                            '<span class="text-sm font-medium truncate" title="' + row.file + '">' + row.file + '</span>' +
                            '<span class="text-xs text-muted-foreground bg-muted px-1.5 py-0.5 rounded flex-shrink-0">' + row.count + 'x</span>' +
                            '</div>';
                    }
                },
                {
                    key: 'locations',
                    label: 'Locations',
                    sortable: false,
                    render: (row) => {
                        const locationsList = row.locationsArray.map(loc =>
                            '<div class="text-xs font-mono text-muted-foreground py-0.5 break-all">' + loc + '</div>'
                        ).join('');
                        return '<div class="space-y-0.5 py-1 max-w-2xl">' + locationsList + '</div>';
                    }
                },
                {
                    key: 'savings',
                    label: 'Savings',
                    align: 'right',
                    width: 'w-32',
                    sortable: true,
                    render: (row) => {
                        return '<span class="text-sm font-semibold text-success bg-success/10 px-2 py-1 rounded-md inline-block">' + row.savingsFormatted + '</span>';
                    }
                }
            ];

            return createDataTable(baseTableId, tableData, columns, {
                pageSize: 10,
                defaultSort: 'savings',
                defaultSortDir: 'desc'
            });
        }

        // Legend configuration - colors and labels
        const LEGEND_ITEMS = [
            { color: '--color-duplicate', label: 'Duplicates' },
            { color: '--color-framework', label: 'Frameworks' },
            { color: '--color-library', label: 'Libraries' },
            { color: '--color-native', label: 'Native Libs' },
            { color: '--color-image', label: 'Images' },
            { color: '--color-video', label: 'Videos' },
            { color: '--color-audio', label: 'Audio' },
            { color: '--color-mlmodel', label: 'ML Models' },
            { color: '--color-dex', label: 'DEX' },
            { color: '--color-asset-catalog', label: 'Asset Catalogs' },
            { color: '--color-font', label: 'Fonts' },
            { color: '--color-localization', label: 'Localization' },
            { color: '--color-resource', label: 'Resources' },
            { color: '--color-ui', label: 'UI' },
            { color: '--color-other', label: 'Other' }
        ];

        // Generate legend items dynamically
        function renderLegend() {
            const container = safeGetElement('legend-container');
            if (!container) return;

            container.textContent = '';

            LEGEND_ITEMS.forEach(item => {
                const legendItem = SafeHTML.createElement('div', {
                    className: 'legend-item',
                    role: 'listitem'
                });

                const colorBox = SafeHTML.createElement('div', {
                    className: 'legend-color',
                    style: { background: 'var(' + item.color + ')' },
                    'aria-hidden': 'true'
                });

                const label = SafeHTML.createElement('span', {
                    className: 'text-xs text-muted-foreground',
                    textContent: item.label
                });

                legendItem.appendChild(colorBox);
                legendItem.appendChild(label);
                container.appendChild(legendItem);
            });
        }

        // Initialize visualizations
        document.addEventListener('DOMContentLoaded', function() {
            // Initialize application state
            AppState.init();
            currentTheme = AppState.theme; // Sync legacy variable
            updateThemeButton();

            // Sync legacy variables with AppState
            originalFileTree = AppState.originalData.fileTree;
            originalCategories = AppState.originalData.categories;
            originalExtensions = AppState.originalData.extensions;

            // Render legend
            renderLegend();

            // Create charts
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
                renderInsights(reportData.optimizations);
            }

            // Set up event delegation
            setupEventDelegation();
        });

        // Event delegation - single listener for all data-action elements
        function setupEventDelegation() {
            document.body.addEventListener('click', function(event) {
                const target = event.target.closest('[data-action]');
                if (!target) return;

                const action = target.getAttribute('data-action');

                switch (action) {
                    case 'toggle-theme':
                        toggleTheme();
                        break;

                    case 'switch-tab':
                        const tabName = target.getAttribute('data-tab');
                        if (tabName) {
                            // Update button styles
                            const tabButtons = document.querySelectorAll('.tab-button');
                            tabButtons.forEach(button => {
                                button.classList.remove('active', 'bg-background', 'text-foreground', 'shadow-sm');
                            });
                            target.classList.add('active', 'bg-background', 'text-foreground', 'shadow-sm');

                            // Update panels
                            const tabPanels = document.querySelectorAll('.tab-panel');
                            tabPanels.forEach(panel => panel.classList.remove('active'));
                            const selectedPanel = safeGetElement(tabName + '-panel');
                            if (selectedPanel) selectedPanel.classList.add('active');

                            // Resize charts
                            if (tabName === 'category') {
                                setTimeout(() => {
                                    ChartFactory.resize('category-chart');
                                    ChartFactory.resize('extension-chart');
                                }, 100);
                            } else if (tabName === 'app-analyzer') {
                                setTimeout(() => {
                                    ChartFactory.resize('treemap');
                                }, 100);
                            }
                        }
                        break;

                    case 'toggle-insight':
                        const index = target.getAttribute('data-index');
                        if (index !== null) {
                            toggleInsight(parseInt(index, 10));
                        }
                        break;

                    case 'toggle-all-insights':
                        toggleAllInsights();
                        break;
                }
            });
        }

        // Keyboard shortcut: Press 'D' to toggle dark mode
        document.addEventListener('keydown', function(event) {
            // Ignore if in an input field
            if (['INPUT', 'TEXTAREA'].includes(event.target.tagName)) {
                return;
            }

            const key = event.key.toLowerCase();

            switch (key) {
                case 'd':
                    event.preventDefault();
                    toggleTheme();
                    break;

                case 'a':
                    event.preventDefault();
                    switchToTab('app-analyzer');
                    break;

                case 'c':
                    event.preventDefault();
                    switchToTab('category');
                    break;

                case 'e':
                    event.preventDefault();
                    toggleAllInsights();
                    break;
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

        // Legacy alias for deepCopy
        const deepCopyNode = TreeUtils.deepCopy;

        // Filter tree based on search query
        function filterTreeByQuery(tree, searchMode, query) {
            if (!tree || searchMode === 'empty') {
                return tree;
            }

            // Backtick mode: find exact node match by name, return it with all children
            if (searchMode === 'backtick') {
                const foundNode = TreeUtils.findByName(tree, query);
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

            // Path-specific mode: only include nodes whose path matches
            if (searchMode === 'path') {
                return TreeUtils.filter(tree, node => {
                    const nodePath = (node.path || '').toLowerCase();
                    const nodeName = node.name.toLowerCase();
                    return nodePath.includes(query) || nodeName.includes(query);
                });
            }

            // Basic mode: match against name, path, fileType
            if (searchMode === 'basic') {
                return TreeUtils.filter(tree, node => {
                    const nodeName = node.name.toLowerCase();
                    const nodePath = (node.path || '').toLowerCase();
                    const nodeType = (node.fileType || '').toLowerCase();
                    return nodeName.includes(query) || nodePath.includes(query) || nodeType.includes(query);
                });
            }

            return tree;
        }

        // Calculate categories and extensions from filtered tree
        function calculateStatsFromTree(tree) {
            const categoryMap = {};
            const extensionMap = {};

            TreeUtils.traverse(tree, node => {
                // Count file types for categories
                if (node.fileType && node.value) {
                    categoryMap[node.fileType] = (categoryMap[node.fileType] || 0) + node.value;
                }

                // Count extensions (only for leaf nodes)
                const isLeaf = !node.children || node.children.length === 0;
                if (isLeaf && node.name && node.value) {
                    const lastDot = node.name.lastIndexOf('.');
                    if (lastDot > 0) {
                        const ext = node.name.substring(lastDot);
                        extensionMap[ext] = (extensionMap[ext] || 0) + node.value;
                    } else {
                        extensionMap['(no ext)'] = (extensionMap['(no ext)'] || 0) + node.value;
                    }
                }
            });

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
            if (filteredTree) {
                ChartFactory.update('treemap', getTreemapOption, filteredTree);
            }
            if (categories && categories.length > 0) {
                ChartFactory.update('category-chart', getCategoryChartOption, categories);
            }
            if (extensions && extensions.length > 0) {
                ChartFactory.update('extension-chart', getExtensionChartOption, extensions);
            }
        }

        // Search functionality with debouncing and caching
        let searchTimeout = null;
        const searchInput = safeGetElement('search-input');
        if (searchInput) searchInput.addEventListener('input', function(e) {
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
                    if (AppState.originalData.fileTree) {
                        updateChartsWithFilteredData(
                            TreeUtils.deepCopy(AppState.originalData.fileTree),
                            AppState.originalData.categories.slice(),
                            AppState.originalData.extensions.slice()
                        );
                    }
                    return;
                }

                // Create cache key
                const cacheKey = parsed.mode + ':' + parsed.query;

                // Check cache first
                let cached = AppState.getCachedSearch(cacheKey);
                if (cached) {
                    updateChartsWithFilteredData(
                        TreeUtils.deepCopy(cached.tree),
                        cached.categories,
                        cached.extensions
                    );
                    return;
                }

                // Filter the tree
                const filteredTree = filterTreeByQuery(
                    TreeUtils.deepCopy(AppState.originalData.fileTree),
                    parsed.mode,
                    parsed.query
                );

                // If no results, show empty state
                if (!filteredTree || (filteredTree.children && filteredTree.children.length === 0)) {
                    const emptyResult = {
                        tree: { name: 'No results', value: 0, children: [] },
                        categories: [],
                        extensions: []
                    };
                    AppState.setCachedSearch(cacheKey, emptyResult);
                    updateChartsWithFilteredData(emptyResult.tree, [], []);
                    return;
                }

                // Recalculate stats from filtered tree
                const stats = calculateStatsFromTree(filteredTree);

                // Cache the result
                AppState.setCachedSearch(cacheKey, {
                    tree: filteredTree,
                    categories: stats.categories,
                    extensions: stats.extensions
                });

                // Update all charts
                updateChartsWithFilteredData(filteredTree, stats.categories, stats.extensions);
            }, 300);
        });
    </script>
</body>
</html>
`
