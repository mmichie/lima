# Claude Instructions for Lima Development

This file contains instructions and context for AI assistants (Claude) working on the Lima project.

## Project Overview

Lima is a beautiful terminal UI for Beancount that provides:
- Smart categorization with ML-powered suggestions
- Interactive transaction management
- Dashboard with charts and reports
- Vim-style keybindings for efficient workflows

**Tech Stack:**
- Language: Go 1.21+
- TUI Framework: Charmbracelet (Bubbletea, Bubbles, Lipgloss)
- Task Management: Beads
- Version Control: Git

## Task Management with Beads

This project uses [Beads](https://github.com/steveyegge/beads) for task tracking. All tasks are stored in `.beads/issues.jsonl` and tracked in git.

### Common Beads Commands

```bash
# View all tasks
bd list

# Filter by priority (0=highest, 4=lowest)
bd list -p 0          # Show P0 tasks only
bd list -p 0,1        # Show P0 and P1 tasks

# Filter by label
bd list -l ui         # UI-related tasks
bd list -l foundation # Foundation tasks

# View task details
bd show lima-2738

# Start working on a task
bd start lima-2738

# Update task status
bd progress lima-2738 -m "Implemented basic navigation"

# Complete a task
bd close lima-2738 -m "Completed TUI framework setup"

# Create new task
bd create "Task title" -d "Description" -p 1 -t feature -l "label1,label2"
```

### Task Priorities

- **P0**: Critical foundation (must be done first)
- **P1**: High priority (core features)
- **P2**: Medium priority (important but not blocking)
- **P3**: Low priority (nice to have, future features)

### Task Types

- `feature`: New functionality
- `task`: General task (refactoring, optimization)
- `bug`: Bug fixes
- `chore`: Maintenance work
- `epic`: Large multi-task effort

## Development Workflow

### Starting a New Task

1. **Find and start a task:**
   ```bash
   bd list -p 0,1           # See high priority tasks
   bd show lima-XXXX        # Review task details
   bd start lima-XXXX       # Mark as in progress
   ```

2. **Create a feature branch:**
   ```bash
   git checkout -b feature/lima-XXXX-short-description
   ```

3. **Implement the feature** following the patterns below

4. **Write tests** for new functionality

5. **Commit with conventional commits:**
   ```bash
   git commit -m "feat(parser): implement transaction parsing (lima-6b23)"
   ```

6. **Update task progress:**
   ```bash
   bd progress lima-XXXX -m "Implemented X, working on Y"
   ```

7. **Complete the task:**
   ```bash
   bd close lima-XXXX -m "Fully implemented and tested"
   ```

### Commit Message Format

Use conventional commits with task reference:

```
<type>(<scope>): <description> (lima-XXXX)

[optional body]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code refactoring
- `test`: Adding tests
- `chore`: Maintenance

**Examples:**
```
feat(ui): add transaction list view (lima-5395)
fix(parser): handle multiline transactions (lima-6b23)
docs: update README with installation instructions
refactor(categorizer): optimize pattern matching (lima-c83d)
test(parser): add unit tests for transaction parsing (lima-2ea5)
```

## Project Structure

```
lima/
├── cmd/lima/              # Main application entry point
│   └── main.go           # Application bootstrap
├── internal/
│   ├── beancount/        # Beancount file parsing
│   ├── categorizer/      # Categorization engine with ML
│   ├── ui/               # Bubbletea views and components
│   │   ├── dashboard/    # Dashboard view
│   │   ├── transactions/ # Transaction list view
│   │   ├── categories/   # Category picker
│   │   └── common/       # Shared UI components
│   ├── pattern/          # Pattern matching and learning
│   ├── reports/          # Report generation
│   └── charts/           # Terminal charts/graphs
├── pkg/
│   ├── parser/           # Shared parsing utilities
│   └── config/           # Configuration management
├── .beads/               # Beads task database
└── go.mod                # Go module definition
```

## Coding Standards

### Go Style

1. **Follow standard Go conventions:**
   - Use `gofmt` for formatting
   - Run `go vet` before committing
   - Follow [Effective Go](https://golang.org/doc/effective_go)

2. **Error handling:**
   ```go
   // Good
   if err != nil {
       return fmt.Errorf("failed to parse transaction: %w", err)
   }

   // Bad - don't ignore errors
   _ = someFunction()
   ```

3. **Naming:**
   - Use camelCase for unexported names
   - Use PascalCase for exported names
   - Interface names should describe behavior (e.g., `Parser`, `Categorizer`)

### Bubbletea Patterns

Each view should follow the Elm architecture:

```go
package transactions

import tea "github.com/charmbracelet/bubbletea"

// Model holds the state
type Model struct {
    transactions []Transaction
    cursor       int
    // ... other state
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
    return nil
}

// Update handles messages and updates state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            return m, tea.Quit
        case "j":
            m.cursor++
        }
    }
    return m, nil
}

// View renders the UI
func (m Model) View() string {
    return "Transaction List"
}
```

### Testing

Write tests for all new functionality:

```go
func TestTransactionParser(t *testing.T) {
    input := `2025-01-01 * "Test"
  Assets:Checking  -100.00 USD
  Expenses:Food     100.00 USD
`
    tx, err := ParseTransaction(input)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if tx.Payee != "Test" {
        t.Errorf("expected payee 'Test', got '%s'", tx.Payee)
    }
}
```

## Working with Bubbletea

### State Management

- Keep state in the model
- Use messages for updates
- Commands for side effects

```go
type Model struct {
    currentView ViewType
    dashboard   dashboard.Model
    transactions transactions.Model
    // ... other views
}

type ViewType int

const (
    DashboardView ViewType = iota
    TransactionsView
    AccountsView
)
```

### Navigation Between Views

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "1":
            m.currentView = DashboardView
            return m, nil
        case "2":
            m.currentView = TransactionsView
            return m, nil
        }
    }

    // Route to current view
    switch m.currentView {
    case DashboardView:
        newDashboard, cmd := m.dashboard.Update(msg)
        m.dashboard = newDashboard.(dashboard.Model)
        return m, cmd
    }

    return m, nil
}
```

## Important Patterns

### Lazy Loading for Large Files

```go
type LazyTransactionList struct {
    file      *os.File
    positions []int64  // File positions for each transaction
    cache     map[int]Transaction
}

func (l *LazyTransactionList) Get(index int) Transaction {
    if tx, ok := l.cache[index]; ok {
        return tx
    }
    // Seek and parse only when needed
    // ...
}
```

### Pattern Matching

```go
type Pattern struct {
    Regex      *regexp.Regexp
    Category   string
    Confidence float64
}

func (p *Pattern) Matches(description string) bool {
    return p.Regex.MatchString(description)
}
```

## Common Pitfalls to Avoid

1. **Don't block the UI thread:**
   - Use `tea.Cmd` for long-running operations
   - Return commands, don't execute directly in Update

2. **Don't mutate models directly:**
   ```go
   // Good
   newModel := m
   newModel.cursor++
   return newModel, nil

   // Bad
   m.cursor++
   return m, nil
   ```

3. **Handle all message types:**
   ```go
   func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       switch msg := msg.(type) {
       case tea.KeyMsg:
           // Handle keys
       case tea.WindowSizeMsg:
           // Handle resize
       default:
           // Always have default case
       }
       return m, nil
   }
   ```

4. **Don't forget error handling:**
   - Always wrap errors with context
   - Use `fmt.Errorf("context: %w", err)` for wrapping

## Performance Considerations

1. **File I/O:**
   - Use buffered readers for large files
   - Implement lazy loading for transaction lists
   - Cache parsed results

2. **Rendering:**
   - Only re-render changed parts when possible
   - Use `lipgloss.JoinVertical` and `JoinHorizontal` efficiently
   - Consider viewport for long lists

3. **Pattern Matching:**
   - Compile regex patterns once, not per match
   - Use map lookups when possible instead of linear search
   - Consider caching categorization results

## Git Workflow

### Before Committing

```bash
# Format code
go fmt ./...

# Run tests
go test ./...

# Check for issues
go vet ./...

# Build to ensure no compile errors
go build ./cmd/lima
```

### Committing

```bash
# Stage changes
git add .

# Commit with conventional commit message
git commit -m "feat(ui): add dashboard view (lima-a2a4)"

# Push to remote
git push origin feature/lima-a2a4-dashboard
```

## Integration with Existing Tools

### Exporting Patterns to Python

When implementing pattern learning, support exporting to Python format for integration with the user's finance repo:

```python
# Export format should match categorize_transactions.py
RULES = {
    r'STARBUCKS': 'Expenses:Food:DiningOut',
    r'SAFEWAY|QFC': 'Expenses:Food:Groceries',
    # ...
}
```

### Config File Location

```
~/.config/lima/config.yaml
```

### Default Beancount File

Support reading from config or command line argument.

## Resources

- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Bubbles Components](https://github.com/charmbracelet/bubbles)
- [Lipgloss Styling](https://github.com/charmbracelet/lipgloss)
- [Beancount Documentation](https://beancount.github.io/)
- [Effective Go](https://golang.org/doc/effective_go)

## Questions or Issues?

- Check existing tasks: `bd list`
- Create new task: `bd create "Your question" -p 2 -t task`
- Review architecture in this file
- Check examples in Charmbracelet repos

---

**Remember:**
- Use beads for ALL tasks
- Follow conventional commits
- Write tests
- Keep the UI responsive
- Make it beautiful with lipgloss
