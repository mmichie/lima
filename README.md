# Lima

> A beautiful terminal UI for Beancount - categorize, report, and analyze your finances from the command line.

Lima is a modern TUI (Terminal User Interface) for [Beancount](https://beancount.github.io/) that makes managing your plain-text accounting ledger fast, efficient, and enjoyable.

## Features

- 🎨 **Beautiful TUI** - Built with [Charmbracelet](https://github.com/charmbracelet) libraries
- 📊 **Comprehensive Dashboard** - Net worth tracking, spending trends, and quick stats
- 💰 **Accounts View** - Hierarchical account tree with balances and recent transactions
- 📝 **Transaction Management** - Browse, search, filter, and categorize transactions
- 🤖 **Smart Categorization** - AI-powered suggestions and pattern learning
- 📈 **Reports & Charts** - Income statements, balance sheets, cash flow, and budgets
- 🔍 **Custom Queries** - Visual query builder and SQL mode
- ⚡ **Fast & Efficient** - Lazy loading, caching, and background indexing
- 🎯 **Vim Keybindings** - Navigate with j/k, search with /, and more

## Installation

### Using Go

```bash
go install github.com/mmichie/lima/cmd/lima@latest
```

### From Source

```bash
git clone https://github.com/mmichie/lima.git
cd lima
go build -o lima ./cmd/lima
sudo mv lima /usr/local/bin/
```

### Using Homebrew (coming soon)

```bash
brew install lima
```

## Quick Start

```bash
# Launch Lima with your Beancount file
lima ~/finance/main.beancount

# Or let Lima use your default file from config
lima

# Start in categorization mode
lima categorize

# Generate a report
lima report income --period "this month"

# Run a custom query
lima query "SELECT * FROM transactions WHERE account ~ 'Expenses:Food'"
```

## Usage

### Main Views

- **Dashboard** (`1`) - Overview of your finances with charts and stats
- **Accounts** (`2`) - Browse your account hierarchy
- **Transactions** (`3`) - View and categorize transactions
- **Reports** (`4`) - Income statements, balance sheets, and more
- **Charts** (`5`) - Visualize spending trends and patterns

### Keyboard Shortcuts

```
Global:
  ?       Help
  q       Quit/Back
  :       Command mode

Transaction View:
  j/k     Navigate down/up
  Enter   Categorize transaction
  Space   Select for batch operations
  c       Categorize selected
  r       Recategorize
  u       Undo
  Ctrl-r  Redo
  /       Search/filter
  f       Toggle filters
  1-9     Quick categorize (recent categories)

Category Picker:
  j/k     Navigate
  h/l     Collapse/expand
  Enter   Select category
  /       Fuzzy search
  Tab     Toggle tree/flat view
  Esc     Cancel
```

## Configuration

Lima looks for configuration in `~/.config/lima/config.yaml`:

```yaml
# Default Beancount file
default_file: ~/finance/main.beancount

# UI preferences
theme: dracula
vim_mode: true
show_help_bar: true

# Dashboard widgets
dashboard:
  - net_worth_chart
  - spending_breakdown
  - recent_transactions
  - budget_status

# Categorization
auto_categorize: true
confidence_threshold: 0.85
learn_patterns: true

# Reporting
default_period: "this_month"
currency: USD
```

## Why Lima?

**Fava is great, but...**
- Requires a browser
- Can be slow for large ledgers
- Not ideal for quick categorization workflows

**Lima gives you:**
- Native terminal experience
- Lightning-fast navigation
- Vim-style efficiency
- Works over SSH
- Perfect for quick edits and categorization
- Beautiful visualizations in your terminal

## Development

```bash
# Clone the repository
git clone https://github.com/mmichie/lima.git
cd lima

# Install dependencies
go mod download

# Run in development mode
go run ./cmd/lima

# Run tests
go test ./...

# Build
go build -o lima ./cmd/lima
```

## Architecture

```
lima/
├── cmd/lima/              # Main application
├── internal/
│   ├── beancount/        # Beancount parser
│   ├── categorizer/      # Categorization engine
│   ├── ui/               # Bubbletea views
│   ├── pattern/          # Pattern matching
│   ├── reports/          # Report generation
│   └── charts/           # Terminal charts
└── pkg/
    ├── parser/           # Shared parsing utilities
    └── config/           # Configuration management
```

## Roadmap

- [ ] Core TUI framework
- [ ] Beancount parser
- [ ] Transaction list view
- [ ] Category picker
- [ ] Smart categorization
- [ ] Dashboard view
- [ ] Reports (Income Statement, Balance Sheet)
- [ ] Charts and visualizations
- [ ] Budget tracking
- [ ] Custom query builder
- [ ] Pattern learning and export
- [ ] Import from CSV/OFX
- [ ] Cloud sync (optional)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

- [Beancount](https://beancount.github.io/) - Amazing plain-text accounting
- [Fava](https://github.com/beancount/fava) - Inspiration for features
- [Charmbracelet](https://github.com/charmbracelet) - Beautiful TUI libraries
- [lima beans](https://en.wikipedia.org/wiki/Lima_bean) - The inspiration for the name

---

Made with ❤️ for the plain-text accounting community
