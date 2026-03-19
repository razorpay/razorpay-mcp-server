#!/usr/bin/env python3
"""
GitHub API-based PR Creator for Log Volume Optimizer

Uses GitHub API directly (via gh) to create branches and PRs without cloning.
Integrates with decision_engine.py for smart optimization decisions.

Usage:
    python github_api_prs.py --report ../reports/target-apps-analysis.json --org razorpay
"""

import argparse
import base64
import json
import os
import re
import subprocess
from datetime import datetime
from typing import Dict, List, Tuple, Optional

# Import decision engine and coralogix integration
try:
    from decision_engine import DecisionEngine, Action
    from coralogix_integration import CoralogixData
except ImportError:
    # Fallback if running from different directory
    import sys
    sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
    from decision_engine import DecisionEngine, Action
    from coralogix_integration import CoralogixData


def run_gh_api(endpoint: str, method: str = "GET", data: dict = None) -> Tuple[bool, dict]:
    """Run a GitHub API call using gh cli."""
    cmd = ['gh', 'api', endpoint]

    if method != "GET":
        cmd.extend(['-X', method])

    if data:
        cmd.extend(['-f', f'data={json.dumps(data)}'])
        # For complex data, use stdin
        input_data = json.dumps(data)
        try:
            result = subprocess.run(
                ['gh', 'api', endpoint, '-X', method, '--input', '-'],
                capture_output=True,
                text=True,
                input=input_data,
                timeout=60
            )
            if result.returncode == 0:
                return True, json.loads(result.stdout) if result.stdout else {}
            return False, {"error": result.stderr}
        except Exception as e:
            return False, {"error": str(e)}

    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
        if result.returncode == 0:
            return True, json.loads(result.stdout) if result.stdout else {}
        return False, {"error": result.stderr}
    except Exception as e:
        return False, {"error": str(e)}


def get_default_branch(org: str, repo: str) -> str:
    """Get the default branch of a repository."""
    cmd = ['gh', 'repo', 'view', f'{org}/{repo}', '--json', 'defaultBranchRef', '-q', '.defaultBranchRef.name']
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0:
            return result.stdout.strip()
    except:
        pass
    return 'master'


def get_branch_sha(org: str, repo: str, branch: str) -> Optional[str]:
    """Get the SHA of a branch."""
    cmd = ['gh', 'api', f'repos/{org}/{repo}/git/ref/heads/{branch}', '-q', '.object.sha']
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0:
            return result.stdout.strip()
    except:
        pass
    return None


def create_branch(org: str, repo: str, branch_name: str, base_sha: str) -> bool:
    """Create a new branch from a base SHA."""
    # First try to delete if exists
    subprocess.run(
        ['gh', 'api', f'repos/{org}/{repo}/git/refs/heads/{branch_name}', '-X', 'DELETE'],
        capture_output=True, timeout=30
    )

    # Create new branch
    cmd = [
        'gh', 'api', f'repos/{org}/{repo}/git/refs',
        '-X', 'POST',
        '-f', f'ref=refs/heads/{branch_name}',
        '-f', f'sha={base_sha}'
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        return result.returncode == 0
    except:
        return False


def get_file_content(org: str, repo: str, path: str, branch: str) -> Tuple[Optional[str], Optional[str]]:
    """Get file content and SHA from GitHub."""
    cmd = ['gh', 'api', f'repos/{org}/{repo}/contents/{path}?ref={branch}']
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=30)
        if result.returncode == 0:
            data = json.loads(result.stdout)
            content = base64.b64decode(data['content']).decode('utf-8')
            return content, data['sha']
    except:
        pass
    return None, None


def update_file(org: str, repo: str, path: str, content: str, sha: str, branch: str, message: str) -> bool:
    """Update a file on GitHub."""
    encoded_content = base64.b64encode(content.encode('utf-8')).decode('utf-8')

    cmd = [
        'gh', 'api', f'repos/{org}/{repo}/contents/{path}',
        '-X', 'PUT',
        '-f', f'message={message}',
        '-f', f'content={encoded_content}',
        '-f', f'sha={sha}',
        '-f', f'branch={branch}'
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
        return result.returncode == 0
    except:
        return False


def create_pr(org: str, repo: str, title: str, body: str, head: str, base: str) -> Tuple[bool, str]:
    """Create a pull request."""
    cmd = [
        'gh', 'pr', 'create',
        '--repo', f'{org}/{repo}',
        '--title', title,
        '--body', body,
        '--head', head,
        '--base', base
    ]
    try:
        result = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
        if result.returncode == 0:
            return True, result.stdout.strip()
        return False, result.stderr
    except Exception as e:
        return False, str(e)


def apply_log_optimization(
    content: str,
    line_num: int,
    language: str = 'auto',
    action: str = 'DOWNGRADE_TO_DEBUG'
) -> Tuple[str, bool, str]:
    """
    Apply log optimization to a specific line based on language and action.

    Args:
        content: File content
        line_num: Line number (1-indexed)
        language: Programming language
        action: Action to take (DOWNGRADE_TO_DEBUG, REMOVE, etc.)

    Returns:
        Tuple of (new_content, was_changed, change_description)
    """
    lines = content.split('\n')
    if line_num < 1 or line_num > len(lines):
        return content, False, ""

    original = lines[line_num - 1]
    new_line = original
    change_desc = ""

    # Auto-detect language from file extension or content patterns
    if language == 'auto':
        if 'Log::' in original or '$this->trace' in original or '$trace->' in original:
            language = 'php'
        elif 'console.' in original or 'logger.' in original.lower():
            language = 'typescript'
        else:
            language = 'go'

    # Handle REMOVE action
    if action == 'REMOVE':
        # Comment out the line instead of deleting (safer)
        indent = len(original) - len(original.lstrip())
        if language == 'go':
            new_line = ' ' * indent + '// ' + original.lstrip() + '  // REMOVED: high-frequency log'
        elif language == 'php':
            new_line = ' ' * indent + '// ' + original.lstrip() + '  // REMOVED: high-frequency log'
        elif language == 'typescript':
            new_line = ' ' * indent + '// ' + original.lstrip() + '  // REMOVED: high-frequency log'
        elif language == 'python':
            new_line = ' ' * indent + '# ' + original.lstrip() + '  # REMOVED: high-frequency log'
        change_desc = "removed (commented out)"

    # Handle DOWNGRADE_TO_DEBUG action
    elif action == 'DOWNGRADE_TO_DEBUG':
        if language == 'go':
            new_line = re.sub(r'\.Info\(', '.Debug(', new_line)
            new_line = re.sub(r'\.Infow\(', '.Debugw(', new_line)
        elif language == 'php':
            new_line = re.sub(r'Log::info\s*\(', 'Log::debug(', new_line)
            new_line = re.sub(r'->info\s*\(', '->debug(', new_line)
            new_line = re.sub(r'\$logger->info\s*\(', '$logger->debug(', new_line)
            new_line = re.sub(r'\$this->logger->info\s*\(', '$this->logger->debug(', new_line)
        elif language == 'typescript':
            new_line = re.sub(r'console\.log\s*\(', 'console.debug(', new_line)
            new_line = re.sub(r'console\.info\s*\(', 'console.debug(', new_line)
            new_line = re.sub(r'logger\.info\s*\(', 'logger.debug(', new_line)
            new_line = re.sub(r'Logger\.info\s*\(', 'Logger.debug(', new_line)
        elif language == 'python':
            new_line = re.sub(r'logging\.info\s*\(', 'logging.debug(', new_line)
            new_line = re.sub(r'logger\.info\s*\(', 'logger.debug(', new_line)
            new_line = re.sub(r'log\.info\s*\(', 'log.debug(', new_line)
        change_desc = "downgraded to DEBUG"

    # Handle SAMPLE action - downgrade to DEBUG (same as DOWNGRADE_TO_DEBUG)
    # Loop logs should be DEBUG level so they don't get shipped to Coralogix
    elif action == 'SAMPLE':
        if language == 'go':
            new_line = re.sub(r'\.Info\(', '.Debug(', new_line)
            new_line = re.sub(r'\.Infow\(', '.Debugw(', new_line)
        elif language == 'php':
            new_line = re.sub(r'Log::info\s*\(', 'Log::debug(', new_line)
            new_line = re.sub(r'->info\s*\(', '->debug(', new_line)
            new_line = re.sub(r'\$logger->info\s*\(', '$logger->debug(', new_line)
            new_line = re.sub(r'\$this->logger->info\s*\(', '$this->logger->debug(', new_line)
        elif language == 'typescript':
            new_line = re.sub(r'console\.log\s*\(', 'console.debug(', new_line)
            new_line = re.sub(r'console\.info\s*\(', 'console.debug(', new_line)
            new_line = re.sub(r'logger\.info\s*\(', 'logger.debug(', new_line)
            new_line = re.sub(r'Logger\.info\s*\(', 'Logger.debug(', new_line)
        elif language == 'python':
            new_line = re.sub(r'logging\.info\s*\(', 'logging.debug(', new_line)
            new_line = re.sub(r'logger\.info\s*\(', 'logger.debug(', new_line)
            new_line = re.sub(r'log\.info\s*\(', 'log.debug(', new_line)
        change_desc = "downgraded to DEBUG (loop log)"

    # For other actions (KEEP, AGGREGATE, USE_METRIC, FLAG_REVIEW), don't modify
    else:
        return content, False, ""

    if new_line != original:
        lines[line_num - 1] = new_line
        return '\n'.join(lines), True, change_desc

    return content, False, ""


def validate_change(original_line: str, new_line: str, action: str, language: str) -> Tuple[bool, str]:
    """
    Validate that a code change is correct before committing.

    Args:
        original_line: The original line of code
        new_line: The modified line of code
        action: The action taken (DOWNGRADE_TO_DEBUG, SAMPLE, REMOVE)
        language: Programming language

    Returns:
        Tuple of (is_valid, error_message)
    """
    # Check 1: Line was actually modified
    if original_line == new_line:
        return False, "Line was not modified"

    # Check 2: For DOWNGRADE/SAMPLE, verify Info was changed to Debug
    if action in ['DOWNGRADE_TO_DEBUG', 'SAMPLE']:
        # Check the log level was actually changed
        if language == 'go':
            if '.Info(' in original_line or '.Infow(' in original_line:
                if '.Debug(' not in new_line and '.Debugw(' not in new_line:
                    return False, "Go: Info was not changed to Debug"
        elif language == 'php':
            if '->info(' in original_line.lower() or 'Log::info(' in original_line:
                if '->debug(' not in new_line.lower() and 'Log::debug(' not in new_line:
                    return False, "PHP: info was not changed to debug"
        elif language == 'typescript':
            if 'console.log(' in original_line or 'console.info(' in original_line or 'logger.info(' in original_line.lower():
                if 'console.debug(' not in new_line and 'logger.debug(' not in new_line.lower():
                    return False, "TypeScript: log/info was not changed to debug"
        elif language == 'python':
            if 'logging.info(' in original_line or 'logger.info(' in original_line:
                if 'logging.debug(' not in new_line and 'logger.debug(' not in new_line:
                    return False, "Python: info was not changed to debug"

    # Check 3: For REMOVE, verify line was commented out
    if action == 'REMOVE':
        if language in ['go', 'php', 'typescript']:
            if not new_line.strip().startswith('//'):
                return False, "Line was not commented out"
        elif language == 'python':
            if not new_line.strip().startswith('#'):
                return False, "Line was not commented out"

    # Check 4: Ensure no random sampling wrapper was added (the old bug)
    if 'rand.Float64()' in new_line or 'Math.random()' in new_line or 'rand(1, 100)' in new_line:
        return False, "ERROR: Random sampling wrapper detected - this is the old bug!"

    # Check 5: Ensure our modification didn't change the bracket balance
    # (multi-line statements legitimately have unbalanced brackets on a single line)
    orig_paren_balance = original_line.count('(') - original_line.count(')')
    new_paren_balance = new_line.count('(') - new_line.count(')')
    if orig_paren_balance != new_paren_balance:
        return False, "Modification changed parentheses balance"

    orig_brace_balance = original_line.count('{') - original_line.count('}')
    new_brace_balance = new_line.count('{') - new_line.count('}')
    if orig_brace_balance != new_brace_balance:
        return False, "Modification changed braces balance"

    return True, ""


def validate_all_changes(changes: List[Dict], original_content: str, new_content: str, language: str) -> Tuple[bool, List[str]]:
    """
    Validate all changes in a file before committing.

    Returns:
        Tuple of (all_valid, list_of_errors)
    """
    errors = []
    original_lines = original_content.split('\n')
    new_lines = new_content.split('\n')

    for change in changes:
        line_num = change['line']
        action = change['action']

        if line_num < 1 or line_num > len(original_lines):
            errors.append(f"Line {line_num}: Invalid line number")
            continue

        # Find the corresponding new line (may have shifted due to removals)
        # For simplicity, check the new content directly
        original_line = original_lines[line_num - 1]

        # Find the new line that corresponds to this change
        # Since we process in reverse order, line numbers should be preserved
        if line_num <= len(new_lines):
            new_line = new_lines[line_num - 1]
            is_valid, error = validate_change(original_line, new_line, action, language)
            if not is_valid:
                errors.append(f"Line {line_num}: {error}")

    return len(errors) == 0, errors


def decide_action_for_log(rec: Dict, decision_engine: DecisionEngine,
                          coralogix_data: CoralogixData = None) -> Tuple[str, str]:
    """
    Use decision engine to determine action for a log recommendation.

    Args:
        rec: Recommendation dict with file, line, level, message, etc.
        decision_engine: DecisionEngine instance
        coralogix_data: Optional CoralogixData for real production frequencies

    Returns:
        Tuple of (action, reason)
    """
    # Infer context from the original action field in the report
    original_action = rec.get('action', '')
    in_loop = 'LOOP' in original_action.upper() or 'CONSOLIDATE' in original_action.upper()

    # Infer error handler from message content
    message = rec.get('message', '').lower()
    in_error_handler = any(kw in message for kw in ['error', 'failed', 'err !=', 'exception'])

    log_entry = {
        'level': rec.get('level', 'INFO'),
        'message': rec.get('message', ''),
        'in_loop': rec.get('in_loop', in_loop),
        'in_error_handler': rec.get('in_error_handler', in_error_handler),
        'file': rec.get('file', ''),
        'line': rec.get('line', 0)
    }

    # Get frequency: prefer real Coralogix data, fall back to estimated daily_units
    frequency = 0
    frequency_source = 'none'

    if coralogix_data and coralogix_data.is_loaded:
        log_message = rec.get('message', '') or rec.get('message_template', '')
        real_freq = coralogix_data.get_daily_frequency(log_message)
        if real_freq > 0:
            frequency = real_freq
            frequency_source = 'coralogix'

    if frequency == 0:
        # Fall back to estimated daily_units from static analysis
        frequency = int(rec.get('daily_units', 0))
        if frequency > 0:
            frequency_source = 'estimated'

    decision = decision_engine.decide(log_entry, frequency=frequency)

    # Annotate reason with frequency source
    reason = decision.reason
    if frequency > 0 and frequency_source == 'coralogix':
        reason = f"[Coralogix: {frequency:,}/day] {reason}"
    elif frequency > 0 and frequency_source == 'estimated':
        reason = f"[Est: {frequency:,}/day] {reason}"

    return decision.action.value, reason


def detect_file_language(file_path: str) -> str:
    """Detect language from file extension."""
    ext = file_path.lower().split('.')[-1] if '.' in file_path else ''
    ext_map = {
        'go': 'go',
        'php': 'php',
        'ts': 'typescript',
        'tsx': 'typescript',
        'js': 'typescript',
        'jsx': 'typescript',
        'py': 'python',
    }
    return ext_map.get(ext, 'go')


def process_repo_via_api(
    org: str,
    repo_name: str,
    recommendations: List[Dict],
    branch_name: str,
    decision_engine: DecisionEngine = None,
    coralogix_data: CoralogixData = None
) -> Dict:
    """Process a repository using GitHub API with smart decision making.

    Args:
        org: GitHub organization
        repo_name: Repository name
        recommendations: List of log recommendations from analysis
        branch_name: Git branch name for changes
        decision_engine: DecisionEngine instance (created if None)
        coralogix_data: Optional CoralogixData with real production frequencies
    """

    if decision_engine is None:
        decision_engine = DecisionEngine()

    result = {
        'repo': repo_name,
        'status': 'pending',
        'changes': 0,
        'savings': 0,
        'pr_url': '',
        'error': '',
        'actions_summary': {}
    }

    print(f"  Getting default branch...")
    default_branch = get_default_branch(org, repo_name)

    print(f"  Getting base SHA from {default_branch}...")
    base_sha = get_branch_sha(org, repo_name, default_branch)
    if not base_sha:
        result['status'] = 'error'
        result['error'] = 'Could not get base branch SHA'
        return result

    print(f"  Creating branch {branch_name}...")
    if not create_branch(org, repo_name, branch_name, base_sha):
        result['status'] = 'error'
        result['error'] = 'Could not create branch'
        return result

    # Apply decision engine to each recommendation
    print(f"  Analyzing {len(recommendations)} logs with decision engine...")
    files_to_update = {}
    actions_summary = {}

    for rec in recommendations:
        full_path = rec.get('file', '')
        line_num = rec.get('line', 0)

        # Extract relative path
        if repo_name in full_path:
            idx = full_path.find(repo_name)
            rel_path = full_path[idx + len(repo_name) + 1:]
        else:
            rel_path = full_path

        # Get decision from engine (uses real Coralogix frequency if available)
        action, reason = decide_action_for_log(rec, decision_engine, coralogix_data)

        # Track action counts
        actions_summary[action] = actions_summary.get(action, 0) + 1

        # Only process actionable items (not KEEP, FLAG_REVIEW, USE_METRIC, AGGREGATE)
        if action not in ['DOWNGRADE_TO_DEBUG', 'REMOVE', 'SAMPLE']:
            continue

        if rel_path not in files_to_update:
            files_to_update[rel_path] = []
        files_to_update[rel_path].append({
            'line': line_num,
            'action': action,
            'reason': reason,
            'savings': rec.get('daily_units', 0),
            'message': rec.get('message', '')[:50]
        })

    print(f"  Decision summary: {actions_summary}")
    result['actions_summary'] = actions_summary

    total_changes = 0
    total_savings = 0
    change_details = []

    for file_path, changes in files_to_update.items():
        print(f"  Processing {file_path} ({len(changes)} changes)...")

        content, sha = get_file_content(org, repo_name, file_path, branch_name)
        if not content:
            print(f"    Could not read file, skipping...")
            continue

        # Detect language from file extension
        file_language = detect_file_language(file_path)

        # Sort changes by line number descending to avoid offset issues
        changes.sort(key=lambda x: x['line'], reverse=True)

        modified = False
        file_changes = []
        skipped_changes = 0
        for change in changes:
            new_content, changed, change_desc = apply_log_optimization(
                content,
                change['line'],
                file_language,
                change['action']
            )
            if changed:
                # Validate each change individually before accepting it
                orig_lines = content.split('\n')
                new_lines = new_content.split('\n')
                line_idx = change['line'] - 1
                if 0 <= line_idx < len(orig_lines) and line_idx < len(new_lines):
                    is_valid, error = validate_change(
                        orig_lines[line_idx],
                        new_lines[line_idx],
                        change['action'],
                        file_language
                    )
                    if not is_valid:
                        skipped_changes += 1
                        if skipped_changes <= 3:  # Show first 3 warnings
                            print(f"    Skipping line {change['line']}: {error}")
                        continue  # Don't apply this change

                content = new_content
                modified = True
                total_changes += 1
                total_savings += change['savings']
                file_changes.append({
                    'line': change['line'],
                    'action': change['action'],
                    'desc': change_desc
                })

        if skipped_changes > 0:
            print(f"    Skipped {skipped_changes} changes that failed validation")

        if modified:
            print(f"    {len(file_changes)} changes validated and applied")

            commit_msg = f"chore: optimize log volume in {file_path}"
            if update_file(org, repo_name, file_path, content, sha, branch_name, commit_msg):
                print(f"    Updated successfully ({len(file_changes)} changes)")
                change_details.extend([{**c, 'file': file_path} for c in file_changes])
                # Get new SHA for subsequent updates
                _, sha = get_file_content(org, repo_name, file_path, branch_name)
            else:
                print(f"    Failed to update")

    if total_changes == 0:
        result['status'] = 'no_changes'
        return result

    result['changes'] = total_changes
    result['savings'] = total_savings

    print(f"  Creating PR ({total_changes} changes, {total_savings:.0f} units savings)...")

    # Build detailed PR body
    downgraded = sum(1 for c in change_details if c['action'] == 'DOWNGRADE_TO_DEBUG')
    removed = sum(1 for c in change_details if c['action'] == 'REMOVE')
    loop_downgraded = sum(1 for c in change_details if c['action'] == 'SAMPLE')
    total_downgraded = downgraded + loop_downgraded

    pr_body = f"""## Log Volume Optimization

Reduces Coralogix log consumption by changing verbose INFO logs to DEBUG level.

### Summary
- **{total_changes} log statements optimized**
- **Estimated savings: {total_savings:,.0f} units/day**

### Changes
| Type | Count | Description |
|------|-------|-------------|
| INFO → DEBUG | {total_downgraded} | Changed log level from INFO to DEBUG |
| REMOVED | {removed} | Commented out redundant logs |

### What Changed
- `logger.Info()` → `logger.Debug()`
- `logger.Infow()` → `logger.Debugw()`

### Decision Rules Applied
Logs were analyzed using these criteria:
- **Preserved (no change)**: ERROR, WARN, FATAL logs
- **Preserved**: Logs in error handlers (`if err != nil`)
- **Preserved**: Business-critical logs (payment, transaction, settlement)
- **Downgraded to DEBUG**: Entry/exit logs (starting, completed, etc.)
- **Downgraded to DEBUG**: High-frequency INFO logs (>1M/day)
- **Downgraded to DEBUG**: Logs inside loops

### Why This Helps
- DEBUG logs are typically **not shipped to Coralogix** in production
- Reduces log volume without losing the ability to enable DEBUG when needed
- All ERROR/WARN logs are preserved for monitoring and alerting

---
Generated by Log Volume Optimizer skill
"""

    success, pr_url = create_pr(
        org, repo_name,
        f"chore({repo_name}): optimize log volume ({total_changes} changes)",
        pr_body,
        branch_name,
        default_branch
    )

    if success:
        result['status'] = 'success'
        result['pr_url'] = pr_url
        print(f"  PR created: {pr_url}")
    else:
        result['status'] = 'pr_failed'
        result['error'] = pr_url[:100]

    return result


def main():
    parser = argparse.ArgumentParser(description='Create PRs using GitHub API with smart decisions')
    parser.add_argument('--report', required=True, help='JSON report from batch_analyze.py')
    parser.add_argument('--org', default='razorpay', help='GitHub organization')
    parser.add_argument('--branch', default='log-volume-optimizer/optimize-logs', help='Branch name')
    parser.add_argument('--repos', help='Comma-separated list of specific repos')
    parser.add_argument('--skip-existing', action='store_true', help='Skip repos with existing PRs')
    parser.add_argument('--output', default='pr-results-api.json', help='Output file')
    parser.add_argument('--dry-run', action='store_true', help='Show what would be done without making changes')
    parser.add_argument('--coralogix-data', help='Path to Coralogix frequency data (JSON/CSV). '
                        'Can be a single file for one app, or a directory with per-app files '
                        'named {app}-coralogix-frequencies.json')
    parser.add_argument('--coralogix-dir', help='Directory with per-app Coralogix data files '
                        '(e.g., reports/ containing pg-router-coralogix-frequencies.json)')
    parser.add_argument('--query-minutes', type=int, default=15,
                        help='Coralogix query time window in minutes (default: 15)')

    args = parser.parse_args()

    # Initialize decision engine
    decision_engine = DecisionEngine()
    print("Decision Engine initialized with rules:")
    print("  - ERROR/FATAL/WARN: Always KEEP")
    print("  - Error handlers: Always KEEP")
    print("  - Business-critical: Always KEEP")
    print("  - Entry/exit patterns: DOWNGRADE to DEBUG")
    print("  - High-frequency INFO: DOWNGRADE or SAMPLE")
    print("  - Metrics patterns: Flag for USE_METRIC")
    print("")

    # Load Coralogix data if provided
    coralogix_single = None  # Single file for all repos
    coralogix_dir = args.coralogix_dir  # Directory with per-app files

    if args.coralogix_data:
        if os.path.isdir(args.coralogix_data):
            coralogix_dir = args.coralogix_data
            print(f"Coralogix data directory: {coralogix_dir}")
        elif os.path.isfile(args.coralogix_data):
            coralogix_single = CoralogixData()
            if args.coralogix_data.endswith('.csv'):
                coralogix_single.load_from_csv(args.coralogix_data,
                                                query_minutes=args.query_minutes)
            else:
                coralogix_single.load_from_json(args.coralogix_data,
                                                 query_minutes=args.query_minutes)
            summary = coralogix_single.get_summary()
            print(f"Coralogix data loaded: {summary['entries']} entries, "
                  f"{summary['total_daily_logs']:,} logs/day, "
                  f"{summary['total_daily_gb']:.2f} GB/day")
        else:
            print(f"WARNING: Coralogix data path not found: {args.coralogix_data}")

    if coralogix_dir:
        print(f"Coralogix per-app data directory: {coralogix_dir}")
    elif not coralogix_single:
        print("NOTE: No Coralogix data provided. Using estimated frequencies only.")
        print("  For better decisions, provide --coralogix-data or --coralogix-dir")
    print("")

    # Load report
    with open(args.report, 'r') as f:
        report = json.load(f)

    repositories = report.get('repositories', [])

    # Filter to specific repos if requested
    if args.repos:
        target_repos = [r.strip() for r in args.repos.split(',')]
        repositories = [r for r in repositories if r.get('repo_name') in target_repos]

    # Filter to only analyzed repos with optimizations
    repositories = [r for r in repositories if r.get('status') == 'analyzed' and r.get('optimization_count', 0) > 0]

    print(f"Processing {len(repositories)} repositories via GitHub API\n")

    if args.dry_run:
        print("DRY RUN MODE - No changes will be made\n")

    results = []

    for i, repo_data in enumerate(repositories, 1):
        repo_name = repo_data.get('repo_name')
        recommendations = repo_data.get('recommendations', [])

        print(f"\n[{i}/{len(repositories)}] {repo_name}")

        if not recommendations:
            print(f"  No recommendations, skipping")
            continue

        print(f"  {len(recommendations)} logs to analyze")

        # Load per-app Coralogix data if available
        repo_coralogix = coralogix_single  # Use single file if provided
        if coralogix_dir:
            # Try to find app-specific data file
            for pattern in [
                f"{repo_name}-coralogix-frequencies.json",
                f"{repo_name}-coralogix.json",
                f"{repo_name}.json",
                f"{repo_name}.csv",
            ]:
                candidate = os.path.join(coralogix_dir, pattern)
                if os.path.isfile(candidate):
                    repo_coralogix = CoralogixData()
                    if candidate.endswith('.csv'):
                        repo_coralogix.load_from_csv(candidate,
                                                      query_minutes=args.query_minutes)
                    else:
                        repo_coralogix.load_from_json(candidate,
                                                       query_minutes=args.query_minutes)
                    break

        if repo_coralogix and repo_coralogix.is_loaded:
            print(f"  Coralogix data: {len(repo_coralogix.entries)} entries loaded")
        else:
            print(f"  Coralogix data: not available (using estimates)")

        # In dry-run mode, just show what would happen
        if args.dry_run:
            action_counts = {}
            freq_sources = {'coralogix': 0, 'estimated': 0, 'none': 0}
            for rec in recommendations:
                action, reason = decide_action_for_log(rec, decision_engine, repo_coralogix)
                action_counts[action] = action_counts.get(action, 0) + 1
                if '[Coralogix:' in reason:
                    freq_sources['coralogix'] += 1
                elif '[Est:' in reason:
                    freq_sources['estimated'] += 1
                else:
                    freq_sources['none'] += 1
            print(f"  Would apply: {action_counts}")
            print(f"  Frequency sources: {freq_sources}")
            continue

        # Process all logs through decision engine with real Coralogix data
        result = process_repo_via_api(
            org=args.org,
            repo_name=repo_name,
            recommendations=recommendations,
            branch_name=args.branch,
            decision_engine=decision_engine,
            coralogix_data=repo_coralogix
        )
        results.append(result)

    if args.dry_run:
        print("\nDry run complete. No changes were made.")
        return

    # Write results
    with open(args.output, 'w') as f:
        json.dump({
            'timestamp': datetime.now().isoformat(),
            'total_repos': len(results),
            'successful': sum(1 for r in results if r['status'] == 'success'),
            'results': results
        }, f, indent=2)

    print(f"\n{'='*50}")
    print("SUMMARY")
    print(f"{'='*50}")

    success_count = sum(1 for r in results if r['status'] == 'success')
    total_changes = sum(r.get('changes', 0) for r in results)
    total_savings = sum(r.get('savings', 0) for r in results)

    print(f"PRs created: {success_count}/{len(results)}")
    print(f"Total changes: {total_changes}")
    print(f"Estimated savings: {total_savings:,.0f} units/day")

    # Print PR URLs
    print("\nPR URLs:")
    for r in results:
        if r.get('pr_url'):
            print(f"  {r['repo']}: {r['pr_url']}")

    print(f"\nResults saved to: {args.output}")


if __name__ == '__main__':
    main()
