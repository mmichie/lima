# Testing Lima

## Quick Start

### 1. Build the application
```bash
go build ./cmd/lima
```

### 2. Run with sample data
```bash
./lima testdata/sample.beancount
```

### 3. Navigate the TUI

**View Switching:**
- `1` - Dashboard view (summary stats)
- `2` - Transactions view (transaction list)
- `3` - Accounts view (account browser)
- `4` - Reports view (coming soon)

**Navigation (in lists):**
- `j` or `↓` - Move down
- `k` or `↑` - Move up
- `g` or `Home` - Jump to top
- `G` or `End` - Jump to bottom
- `Ctrl+f` or `PgDn` - Page down
- `Ctrl+b` or `PgUp` - Page up

**General:**
- `q` or `Ctrl+c` - Quit
- `?` - Help (not yet implemented)

## Running Tests

### Run all tests
```bash
go test ./... -v
```

### Run specific package tests
```bash
# Parser tests
go test ./internal/beancount/... -v

# UI tests
go test ./internal/ui/... -v

# Config tests
go test ./pkg/config/... -v
```

### Run with coverage
```bash
go test ./... -cover
```

## Testing with Your Own Beancount File

### Option 1: Command line argument
```bash
./lima /path/to/your/ledger.beancount
```

### Option 2: Configuration file
1. Create config directory:
   ```bash
   mkdir -p ~/.config/lima
   ```

2. Copy example config:
   ```bash
   cp config.example.yaml ~/.config/lima/config.yaml
   ```

3. Edit config to set your default ledger:
   ```yaml
   files:
     default_ledger: /path/to/your/ledger.beancount
   ```

4. Run without arguments:
   ```bash
   ./lima
   ```

## Demo Utility

Test the parser directly without the TUI:

```bash
go run ./cmd/parser-demo testdata/sample.beancount
```

This displays:
- Total transaction count
- Account list
- Commodity list
- First 5 transactions with details

## What to Look For

### Dashboard View (Press `1`)
- Total transaction count
- Account count
- Commodity count
- Recent transactions list

### Transactions View (Press `2`)
- Scrollable list of all transactions
- Date, flag, payee, account, amount
- Selected item highlighted
- Navigation counter (e.g., "5/7")

### Accounts View (Press `3`)
- Accounts grouped by type:
  - Assets
  - Liabilities
  - Equity
  - Income
  - Expenses
- Selected account highlighted

## Sample Data

The included `testdata/sample.beancount` contains:
- 7 transactions
- 7 accounts (Assets, Income, Expenses)
- Mix of cleared (*) and pending (!) transactions
- Tags and links
- Transaction metadata
- Date range: 2025-01-01 to 2025-01-25

## Customization

### Theme Colors
Edit `~/.config/lima/config.yaml`:
```yaml
theme:
  primary: "#00D9FF"    # Cyan
  success: "#00FF00"    # Green
  error: "#FF0000"      # Red
```

### Keybindings
Edit `~/.config/lima/config.yaml`:
```yaml
keybindings:
  quit: ["q", "ctrl+c"]
  up: ["up", "k"]
  down: ["down", "j"]
```

### UI Preferences
```yaml
ui:
  default_view: dashboard  # or: transactions, accounts, reports
  page_size: 20
  date_format: "2006-01-02"
```

## Troubleshooting

### File not found
```
Error opening file: failed to open file: open testdata/sample.beancount: no such file or directory
```
**Solution:** Run from the project root directory or provide absolute path.

### Parse error
```
Error opening file: failed to build index: ...
```
**Solution:** Check that your Beancount file has valid syntax.

### Display issues
If colors don't display correctly:
- Ensure your terminal supports 256 colors
- Try different terminals (iTerm2, Alacritty, etc.)
- Check theme settings in config

## Performance Testing

Test with large files:

1. Create a large Beancount file (10k+ transactions)
2. Open with Lima
3. Navigate through transactions - should be instant (lazy loading)
4. Memory usage should remain constant

## Next Steps

After testing basic functionality, try:
1. Creating your own config file
2. Testing with your real Beancount ledger
3. Exploring different views and navigation
4. Customizing theme colors and keybindings
