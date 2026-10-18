# Resumator

A desktop application to build and maintain multiple versions of a professional
resume. Built with [Wails v2](https://wails.io) (Go) and Svelte 5.

## Features

- Edit each resume section independently: contact, skills, professional
  experience, education, plus free-form custom sections.
- Per-preset selection: show or hide any section, entry, or bullet point.
- Live preview of the resume on the right while you edit on the left.
- Presets: save the current selection under a new name to maintain several
  resume variants (for example a short and a detailed version).
- Export the current preset as a PDF.
- Light and dark themes.
- All data is stored locally in SQLite.

## Data location

The SQLite database is created automatically under the user config directory:

- Linux: `~/.config/resumator/resumator.db`
- Windows: `%APPDATA%\resumator\resumator.db`
- macOS: `~/Library/Application Support/resumator/resumator.db`

## Development

Prerequisites: Go 1.25+, Node.js, and the
[Wails CLI](https://wails.io/docs/gettingstarted/installation)
(`go install github.com/wailsapp/wails/v2/cmd/wails@latest`). On Linux, the
webkit2gtk development packages are required (see Wails docs for the list).

Run in live development mode with hot reload:

```
wails dev
```

On Linux, Wails' dev server only serves trusted hosts, and the embedded
webview loads the app from `http://wails.localhost:34115`. Allow that host
or the window stays black:

```
WAILS_DEV_ALLOWED_HOSTS=wails.localhost wails dev -tags webkit2_41
```

(Alternatively `export WAILS_DEV_ALLOWED_HOSTS=wails.localhost` once in your
shell profile.) This only affects `wails dev`; production builds are not
affected.

Type-check the frontend:

```
cd frontend && npm run check
```

## Building

Build a redistributable for your platform:

```
wails build
```

Cross-platform builds are performed on the target platform (Wails relies on
the native webview: WebView2 on Windows, WKWebView on macOS, webkit2gtk on
Linux):

- **Windows**: `wails build` (or `wails build -nsis` for an installer)
- **macOS**: `wails build` (or `wails build -platform darwin/universal`)
- **Linux**: `wails build`

On Linux systems that only ship `webkit2gtk-4.1` (for example recent Fedora
releases), add the matching build tag:

```
wails build -tags webkit2_41
```

The built binary is placed in `build/bin/`.

## Architecture

```
frontend/                 Svelte 5 + TypeScript UI
  src/lib/stores/         Central reactive state ($state), debounced autosave
  src/components/         Sidebar, section editors, live A4 preview
internal/store/           SQLite persistence (modernc.org/sqlite, no CGO)
internal/pdf/             PDF generation (maroto v2)
app.go                    Wails-bound API surface
```

Section and entry selections live in preset-scoped tables: a preset stores
which sections are visible and which items/details are included, while the
underlying content is shared across presets. Editing a section updates every
resume variant; toggling entries only affects the active preset.
