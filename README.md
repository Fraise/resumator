🤖 **Disclaimer:** this application was vibe coded using AI. It works, but
treat the code with appropriate curiosity before trusting it with anything
critical.

# Resumator

<p align="center">
	<img src="logo.png" alt="Resumator logo" width="300"/>
</p>

Resumator is a desktop application for building and maintaining multiple
versions of a professional resume. You write your content once — contact
details, skills, professional experience, education, and any number of
custom sections — then decide, per resume variant, which sections, entries
and bullet points appear on it.

The left pane edits the resume; the right pane shows a live, paginated
preview of the result on A4 or US Letter paper. The preview and the PDF
export share the same layout engine, so what you see is what gets printed:
sections are never cut in half across pages, and page breaks land in the
same place on screen and on paper.

Everything is stored locally in a single SQLite database — no accounts, no
cloud, no telemetry. Built with [Wails v2](https://wails.io) (Go backend,
Svelte 5 frontend) for Linux, Windows and macOS.

## Features

- Edit each resume section independently: contact, skills, professional
  experience, education, plus free-form custom sections (with drag-and-drop
  reordering everywhere).
- Per-preset selection: show or hide any section, entry, or bullet point.
- Presets: save the current selection under a new name to maintain several
  resume variants (for example a short and a detailed version).
- Live paginated preview on A4 or US Letter, matching the PDF export.
- Export the current preset as a PDF, with optional per-section title and
  separator colors.
- Light and dark themes.

## Running the app

Prebuilt binaries for Linux, Windows and macOS are attached to each
[GitHub release](../../releases) automatically (a workflow builds all three
platforms whenever a release is published). Download the archive for your
platform, unpack it, and run the binary.

To build and run it yourself instead:

```
wails build
```

The executable is placed in `build/bin/` (`resumator.exe` on Windows,
`resumator.app` on macOS).

### Where your data is saved

All resume content, presets and selections live in a single SQLite file
created automatically on first launch:

- **Linux:** `~/.config/resumator/resumator.db`
- **Windows:** `%APPDATA%\resumator\resumator.db`
- **macOS:** `~/Library/Application Support/resumator/resumator.db`

Back up or copy that file to move your resumes to another machine. A few UI
preferences (theme, page format, pane width) are stored in the app's local
storage instead of the database.

## Developing the app

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
affected. The `-tags webkit2_41` part is needed on distros that only ship
`webkit2gtk-4.1` (for example recent Fedora releases); on distros with
`webkit2gtk-4.0`, plain `wails dev` is enough.

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

- **Windows:** `wails build` (or `wails build -nsis` for an installer)
- **macOS:** `wails build` (or `wails build -platform darwin/universal`)
- **Linux:** `wails build` (add `-tags webkit2_41` on webkit2gtk-4.1-only
  distros)

Release binaries for all three platforms are built automatically by the
GitHub Actions workflow in `.github/workflows/release.yml` whenever a
release is published, and attached to that release.

## Architecture

```
frontend/                 Svelte 5 + TypeScript UI
  src/lib/stores/         Central reactive state ($state), debounced autosave
  src/components/         Sidebar, section editors, paginated live preview
internal/store/           SQLite persistence (modernc.org/sqlite, no CGO)
internal/pdf/             PDF generation + page layout engine (maroto v2)
app.go                    Wails-bound API surface
```

Section and entry selections live in preset-scoped tables: a preset stores
which sections are visible and which items/details are included, while the
underlying content is shared across presets. Editing a section updates every
resume variant; toggling entries only affects the active preset. The page
layout (which section starts which page) is computed once by the Go layout
engine and shared by the preview and the PDF export, so the two always agree.
