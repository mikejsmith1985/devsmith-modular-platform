#!/bin/bash
# Recovery Helper Script
# Helps recover work after a Claude Code crash
# Usage: .claude/hooks/recovery-helper.sh [--list|--restore|--status]

set -euo pipefail

RECOVERY_BRANCH_PREFIX="claude-recovery"
LOG_DIR=".claude/recovery-logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}================================${NC}"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Show recovery status
show_status() {
    print_header "Recovery Status Check"

    echo ""
    echo "Current Branch: $(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'Not in git repo')"
    echo "Working Directory: $(pwd)"
    echo ""

    # Check for recovery branches
    echo "Available Recovery Branches:"
    recovery_branches=$(git branch --list "${RECOVERY_BRANCH_PREFIX}-*" 2>/dev/null || echo "")
    if [ -z "$recovery_branches" ]; then
        print_warning "No recovery branches found"
    else
        echo "$recovery_branches" | while read -r branch; do
            branch_name=$(echo "$branch" | sed 's/^[* ]*//')
            last_commit=$(git log -1 --format="%cr - %s" "$branch_name" 2>/dev/null || echo "unknown")
            print_success "$branch_name"
            echo "   Last commit: $last_commit"
        done
    fi

    echo ""
    echo "Recent Session Logs:"
    if [ -d "$LOG_DIR" ]; then
        recent_logs=$(find "$LOG_DIR" -name "session-*.md" -mtime -7 | sort -r | head -5)
        if [ -z "$recent_logs" ]; then
            print_warning "No recent session logs found"
        else
            echo "$recent_logs" | while read -r logfile; do
                basename_log=$(basename "$logfile")
                line_count=$(wc -l < "$logfile")
                print_success "$basename_log ($line_count lines)"
            done
        fi
    else
        print_warning "No log directory found"
    fi

    echo ""
    echo "Todo List Status:"
    if [ -f ".claude/todos.json" ]; then
        print_success "Todo list found (.claude/todos.json)"
        # Count pending/in-progress tasks
        in_progress=$(grep -c '"status":"in_progress"' .claude/todos.json 2>/dev/null || echo "0")
        pending=$(grep -c '"status":"pending"' .claude/todos.json 2>/dev/null || echo "0")
        completed=$(grep -c '"status":"completed"' .claude/todos.json 2>/dev/null || echo "0")
        echo "   In Progress: $in_progress | Pending: $pending | Completed: $completed"
    else
        print_warning "No todo list found"
    fi
}

# List recent session logs
list_logs() {
    print_header "Recent Session Logs"

    if [ ! -d "$LOG_DIR" ]; then
        print_error "No log directory found at $LOG_DIR"
        exit 1
    fi

    recent_logs=$(find "$LOG_DIR" -name "session-*.md" -mtime -7 | sort -r)

    if [ -z "$recent_logs" ]; then
        print_warning "No session logs found in the last 7 days"
        exit 0
    fi

    echo ""
    echo "Select a log to view:"
    echo ""

    select logfile in $recent_logs "Cancel"; do
        if [ "$logfile" = "Cancel" ]; then
            echo "Cancelled"
            exit 0
        elif [ -n "$logfile" ]; then
            echo ""
            print_header "Contents of $(basename $logfile)"
            echo ""
            cat "$logfile"
            exit 0
        fi
    done
}

# Restore from recovery branch
restore_from_recovery() {
    print_header "Restore from Recovery Branch"

    current_branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")
    if [ -z "$current_branch" ]; then
        print_error "Not in a git repository"
        exit 1
    fi

    # List available recovery branches
    recovery_branches=$(git branch --list "${RECOVERY_BRANCH_PREFIX}-*" 2>/dev/null | sed 's/^[* ]*//')

    if [ -z "$recovery_branches" ]; then
        print_error "No recovery branches found"
        exit 1
    fi

    echo ""
    echo "Current branch: $current_branch"
    echo ""
    echo "Available recovery branches:"
    echo ""

    select branch in $recovery_branches "Cancel"; do
        if [ "$branch" = "Cancel" ]; then
            echo "Cancelled"
            exit 0
        elif [ -n "$branch" ]; then
            echo ""
            print_header "Recovery Plan"
            echo ""
            echo "This will cherry-pick commits from: $branch"
            echo "Into current branch: $current_branch"
            echo ""

            # Show what commits would be applied
            echo "Commits to be applied:"
            git log --oneline "$current_branch..$branch" 2>/dev/null || {
                print_error "No commits to apply from $branch"
                exit 1
            }

            echo ""
            read -p "Proceed with recovery? (y/n) " -n 1 -r
            echo ""

            if [[ $REPLY =~ ^[Yy]$ ]]; then
                # Cherry-pick commits
                if git cherry-pick "$current_branch..$branch"; then
                    print_success "Recovery successful!"
                    echo ""
                    echo "Your work has been restored to branch: $current_branch"
                else
                    print_error "Cherry-pick failed - there may be conflicts"
                    echo ""
                    echo "You can resolve conflicts manually and run:"
                    echo "  git cherry-pick --continue"
                    echo "Or abort with:"
                    echo "  git cherry-pick --abort"
                    exit 1
                fi
            else
                echo "Recovery cancelled"
            fi
            exit 0
        fi
    done
}

# Main script
case "${1:-status}" in
    --status|status)
        show_status
        ;;
    --list|list)
        list_logs
        ;;
    --restore|restore)
        restore_from_recovery
        ;;
    --help|help|-h)
        echo "Claude Code Recovery Helper"
        echo ""
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  status    Show recovery status (default)"
        echo "  list      List and view recent session logs"
        echo "  restore   Restore work from a recovery branch"
        echo "  help      Show this help message"
        ;;
    *)
        print_error "Unknown command: $1"
        echo "Run '$0 --help' for usage"
        exit 1
        ;;
esac
