#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Credential Logging Security Scanner ===${NC}"
echo "Scanning for potential credential logging issues in the codebase..."
echo

# Function to check if a file should be excluded
should_exclude() {
    local file="$1"
    # Skip test files, vendor directories, and documentation
    [[ "$file" =~ \.git/ ]] || \
    [[ "$file" =~ /vendor/ ]] || \
    [[ "$file" =~ _test\.go$ ]] || \
    [[ "$file" =~ \.md$ ]] || \
    [[ "$file" =~ \.txt$ ]] || \
    [[ "$file" =~ \.yml$ ]] || \
    [[ "$file" =~ \.yaml$ ]] || \
    [[ "$file" =~ \.json$ ]] || \
    [[ "$file" =~ /bin/ ]] || \
    [[ "$file" =~ /\.cursor/ ]]
}

# Function to check for credential logging patterns
check_credential_logging() {
    local issues_found=0
    
    echo -e "${YELLOW}1. Checking for fmt.Printf with sensitive credentials...${NC}"
    
    # Find fmt.Printf patterns with potential credentials
    if command -v rg &> /dev/null; then
        SEARCH_CMD="rg"
    elif command -v grep &> /dev/null; then
        SEARCH_CMD="grep -r"
    else
        echo -e "${RED}Error: Neither ripgrep nor grep found${NC}"
        exit 1
    fi
    
    # Patterns to search for
    patterns=(
        'fmt\.Printf.*[Aa]uth'
        'fmt\.Printf.*[Pp]assword'
        'fmt\.Printf.*[Tt]oken'
        'fmt\.Printf.*[Kk]ey'
        'fmt\.Printf.*[Ss]ecret'
        'fmt\.Printf.*[Cc]redential'
        'fmt\.Printf.*[Bb]asic[Aa]uth'
        'fmt\.Printf.*[Aa]pi[Kk]ey'
        'fmt\.Printf.*[Aa]uthorization'
        'fmt\.Printf.*[Hh]eader.*[Aa]uth'
        'fmt\.Printf.*[Hh]eader.*[Tt]oken'
        'fmt\.Printf.*[Hh]eader.*[Kk]ey'
        'fmt\.Printf.*[Hh]eader.*[Pp]assword'
        'fmt\.Printf.*[Hh]eader.*[Ss]ecret'
        'fmt\.Printf.*[Hh]eader.*[Cc]redential'
        'fmt\.Sprintf.*[Aa]uth'
        'fmt\.Sprintf.*[Pp]assword'
        'fmt\.Sprintf.*[Tt]oken'
        'fmt\.Sprintf.*[Kk]ey'
        'fmt\.Sprintf.*[Ss]ecret'
        'fmt\.Sprintf.*[Cc]redential'
        'log\.Printf.*[Aa]uth'
        'log\.Printf.*[Pp]assword'
        'log\.Printf.*[Tt]oken'
        'log\.Printf.*[Kk]ey'
        'log\.Printf.*[Ss]ecret'
        'log\.Printf.*[Cc]redential'
        'print.*[Aa]uth.*:'
        'print.*[Pp]assword.*:'
        'print.*[Tt]oken.*:'
        'print.*[Kk]ey.*:'
        'print.*[Ss]ecret.*:'
        'print.*[Cc]redential.*:'
    )
    
    for pattern in "${patterns[@]}"; do
        if [[ "$SEARCH_CMD" == "rg" ]]; then
            results=$(rg -i "$pattern" --type go --line-number --no-heading . 2>/dev/null || true)
        else
            results=$(grep -rn -i "$pattern" --include="*.go" . 2>/dev/null || true)
        fi
        
        if [[ -n "$results" ]]; then
            echo -e "${RED}SECURITY ISSUE - Pattern: $pattern${NC}"
            while IFS= read -r line; do
                file=$(echo "$line" | cut -d':' -f1)
                line_num=$(echo "$line" | cut -d':' -f2)
                content=$(echo "$line" | cut -d':' -f3-)
                
                if ! should_exclude "$file"; then
                    echo -e "  ${YELLOW}File:${NC} $file:$line_num"
                    echo -e "  ${RED}Issue:${NC} $content"
                    echo
                    ((issues_found++))
                fi
            done <<< "$results"
        fi
    done
    
    echo -e "${YELLOW}2. Checking for hardcoded credentials...${NC}"
    
    # Check for hardcoded credentials
    hardcoded_patterns=(
        'password.*=.*".*"'
        'token.*=.*".*"'
        'api_key.*=.*".*"'
        'secret.*=.*".*"'
        'basic_auth.*=.*".*"'
        'Authorization.*=.*".*"'
        'Bearer.*".*"'
        'Basic.*".*"'
    )
    
    for pattern in "${hardcoded_patterns[@]}"; do
        if [[ "$SEARCH_CMD" == "rg" ]]; then
            results=$(rg -i "$pattern" --type go --line-number --no-heading . 2>/dev/null || true)
        else
            results=$(grep -rn -i "$pattern" --include="*.go" . 2>/dev/null || true)
        fi
        
        if [[ -n "$results" ]]; then
            echo -e "${RED}POTENTIAL HARDCODED CREDENTIAL - Pattern: $pattern${NC}"
            while IFS= read -r line; do
                file=$(echo "$line" | cut -d':' -f1)
                line_num=$(echo "$line" | cut -d':' -f2)
                content=$(echo "$line" | cut -d':' -f3-)
                
                if ! should_exclude "$file"; then
                    echo -e "  ${YELLOW}File:${NC} $file:$line_num"
                    echo -e "  ${RED}Issue:${NC} $content"
                    echo
                    ((issues_found++))
                fi
            done <<< "$results"
        fi
    done
    
    echo -e "${YELLOW}3. Checking for environment variable logging...${NC}"
    
    # Check for environment variable logging that might expose credentials
    env_patterns=(
        'fmt\.Printf.*os\.Getenv.*[Aa]uth'
        'fmt\.Printf.*os\.Getenv.*[Pp]assword'
        'fmt\.Printf.*os\.Getenv.*[Tt]oken'
        'fmt\.Printf.*os\.Getenv.*[Kk]ey'
        'fmt\.Printf.*os\.Getenv.*[Ss]ecret'
        'fmt\.Printf.*os\.Getenv.*[Cc]redential'
        'log\.Printf.*os\.Getenv.*[Aa]uth'
        'log\.Printf.*os\.Getenv.*[Pp]assword'
        'log\.Printf.*os\.Getenv.*[Tt]oken'
        'log\.Printf.*os\.Getenv.*[Kk]ey'
        'log\.Printf.*os\.Getenv.*[Ss]ecret'
        'log\.Printf.*os\.Getenv.*[Cc]redential'
    )
    
    for pattern in "${env_patterns[@]}"; do
        if [[ "$SEARCH_CMD" == "rg" ]]; then
            results=$(rg -i "$pattern" --type go --line-number --no-heading . 2>/dev/null || true)
        else
            results=$(grep -rn -i "$pattern" --include="*.go" . 2>/dev/null || true)
        fi
        
        if [[ -n "$results" ]]; then
            echo -e "${RED}ENVIRONMENT VARIABLE LOGGING - Pattern: $pattern${NC}"
            while IFS= read -r line; do
                file=$(echo "$line" | cut -d':' -f1)
                line_num=$(echo "$line" | cut -d':' -f2)
                content=$(echo "$line" | cut -d':' -f3-)
                
                if ! should_exclude "$file"; then
                    echo -e "  ${YELLOW}File:${NC} $file:$line_num"
                    echo -e "  ${RED}Issue:${NC} $content"
                    echo
                    ((issues_found++))
                fi
            done <<< "$results"
        fi
    done
    
    echo -e "${YELLOW}4. Checking for improper logging frameworks usage...${NC}"
    
    # Check for fmt.Printf instead of proper logging frameworks
    printf_results=$(find . -name "*.go" -not -path "./vendor/*" -not -path "./.git/*" -not -name "*_test.go" -exec grep -l "fmt\.Printf" {} \; 2>/dev/null || true)
    
    if [[ -n "$printf_results" ]]; then
        echo -e "${YELLOW}Files using fmt.Printf instead of proper logging framework:${NC}"
        while IFS= read -r file; do
            if [[ -n "$file" ]] && ! should_exclude "$file"; then
                count=$(grep -c "fmt\.Printf" "$file" 2>/dev/null || echo "0")
                echo -e "  ${YELLOW}File:${NC} $file (${count} occurrences)"
                # Show the actual lines
                grep -n "fmt\.Printf" "$file" | head -3 | while IFS= read -r line; do
                    echo -e "    ${NC}$line"
                done
                if [[ $(grep -c "fmt\.Printf" "$file" 2>/dev/null || echo "0") -gt 3 ]]; then
                    echo -e "    ${YELLOW}... and $((count - 3)) more${NC}"
                fi
                echo
                ((issues_found++))
            fi
        done <<< "$printf_results"
    fi
    
    return $issues_found
}

# Function to provide recommendations
provide_recommendations() {
    echo -e "${GREEN}=== RECOMMENDATIONS ===${NC}"
    echo
    echo -e "${YELLOW}1. Replace fmt.Printf with proper logging framework:${NC}"
    echo "   - Use logrus.WithFields() for structured logging"
    echo "   - Use appropriate log levels (Debug, Info, Warn, Error)"
    echo "   - Example: logger.WithFields(logrus.Fields{\"key\": \"value\"}).Info(\"message\")"
    echo
    echo -e "${YELLOW}2. Redact sensitive information:${NC}"
    echo "   - Never log passwords, tokens, or API keys directly"
    echo "   - Use redaction functions like: redactField(sensitiveValue)"
    echo "   - Log only non-sensitive metadata (e.g., \"has_auth\": true)"
    echo
    echo -e "${YELLOW}3. Secure header logging:${NC}"
    echo "   - Always redact Authorization headers"
    echo "   - Use helper functions to identify sensitive headers"
    echo "   - Log header names but redact values for sensitive headers"
    echo
    echo -e "${YELLOW}4. Environment variable safety:${NC}"
    echo "   - Never log environment variables containing credentials"
    echo "   - Log only boolean flags like \"has_env_auth\": true"
    echo
    echo -e "${YELLOW}5. Configuration logging:${NC}"
    echo "   - Log configuration parameters but redact sensitive values"
    echo "   - Use partial redaction for usernames (show first 3 chars + ***)"
    echo "   - Completely redact passwords and tokens"
    echo
}

# Main execution
main() {
    local total_issues=0
    
    # Check if we're in a Go project
    if [[ ! -f "go.mod" ]]; then
        echo -e "${RED}Error: Not in a Go project directory (go.mod not found)${NC}"
        exit 1
    fi
    
    # Run the credential logging check
    check_credential_logging
    total_issues=$?
    
    echo -e "${GREEN}=== SCAN COMPLETE ===${NC}"
    echo
    
    if [[ $total_issues -eq 0 ]]; then
        echo -e "${GREEN}✅ No credential logging issues found!${NC}"
    else
        echo -e "${RED}❌ Found $total_issues potential credential logging issues${NC}"
        echo
        provide_recommendations
    fi
    
    return $total_issues
}

# Run the main function
main "$@" 