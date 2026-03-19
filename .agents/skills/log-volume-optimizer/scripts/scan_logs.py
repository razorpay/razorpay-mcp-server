#!/usr/bin/env python3
"""
Log Scanner for Log Volume Optimizer Skill

Scans a repository for log statements and extracts metadata for analysis.
Supports Go, Python, and Java codebases.

Usage:
    python scan_logs.py /path/to/repo --language go --output logs.json
"""

import argparse
import json
import os
import re
from dataclasses import asdict, dataclass
from pathlib import Path
from typing import Dict, List, Optional


@dataclass
class LogStatement:
    """Represents a detected log statement."""
    file: str
    line: int
    column: int
    level: str
    message: str
    function: str
    statement: str
    in_loop: bool
    in_error_handler: bool
    language: str
    estimated_size_bytes: int = 0
    message_template: str = ""  # Normalized message for matching to Coralogix

    def to_dict(self) -> Dict:
        return asdict(self)


# Directories to always exclude (from Rajeev's approach)
EXCLUDED_DIRS = [
    'vendor', 'build', 'config', 'bin', 'scripts', 'templates',
    'node_modules', '.git', 'dist', 'coverage', '__pycache__',
    'venv', '.venv', 'testdata', 'mocks', 'mock'
]


def extract_message_template(message: str) -> str:
    """
    Extract a normalized message template for matching to Coralogix data.

    Replaces dynamic values with placeholders:
    - UUIDs, IDs, numbers -> {id}
    - Email addresses -> {email}
    - URLs -> {url}
    - IP addresses -> {ip}
    - Timestamps -> {time}
    - Generic values in structured logs -> {value}

    Args:
        message: Raw log message

    Returns:
        Normalized template string
    """
    if not message:
        return ""

    template = message

    # Replace UUIDs
    template = re.sub(
        r'[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}',
        '{uuid}', template
    )

    # Replace hex IDs (like MongoDB ObjectIds)
    template = re.sub(r'[0-9a-fA-F]{24}', '{id}', template)

    # Replace email addresses
    template = re.sub(r'[\w\.-]+@[\w\.-]+\.\w+', '{email}', template)

    # Replace URLs
    template = re.sub(r'https?://[^\s\'"]+', '{url}', template)

    # Replace IP addresses
    template = re.sub(r'\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}', '{ip}', template)

    # Replace timestamps (various formats)
    template = re.sub(r'\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[.\d]*Z?', '{time}', template)

    # Replace numbers (but keep short ones that might be meaningful)
    template = re.sub(r'\b\d{6,}\b', '{id}', template)  # Long numbers are likely IDs

    # Replace Go format verbs
    template = re.sub(r'%[+#\-0-9.]*[vTtbcdoqxXUeEfFgGsp]', '{value}', template)

    # Replace interpolation placeholders in various languages
    template = re.sub(r'\$\{[^}]+\}', '{value}', template)  # ${var}
    template = re.sub(r'\{\{[^}]+\}\}', '{value}', template)  # {{var}}

    # Normalize whitespace
    template = ' '.join(template.split())

    # Truncate to first 100 chars for matching
    return template[:100]


class GoLogScanner:
    """Scans Go source files for log statements."""
    
    # Common Go logging patterns
    LOG_PATTERNS = [
        # Standard library
        r'log\.(Print|Printf|Println|Fatal|Fatalf|Fatalln|Panic|Panicf|Panicln)\s*\(',
        # Logrus
        r'(log|logger|lgr)\.(Debug|Info|Warn|Warning|Error|Fatal|Panic)(f|ln)?\s*\(',
        # Zap
        r'(log|logger|lgr)\.(Debug|Info|Warn|Error|DPanic|Panic|Fatal)(w)?\s*\(',
        # Razorpay chained pattern
        r'(logger|lgr)\.(Log|Logger)\s*\([^)]*\)\.(Debug|Info|Warn|Error|Fatal)(w)?\s*\(',
        # With fields pattern
        r'(logger|lgr)\.(Log|Logger)\s*\([^)]*\)\.With[^.]*\.(Debug|Info|Warn|Error|Fatal)(w)?\s*\(',
    ]
    
    LOOP_PATTERNS = [
        r'\bfor\s+',
        r'\brange\s+',
    ]
    
    ERROR_HANDLER_PATTERNS = [
        r'if\s+err\s*!=\s*nil',
        r'if\s+.*err.*\{',
    ]
    
    FUNCTION_PATTERN = r'func\s+(?:\([^)]+\)\s+)?(\w+)\s*\('
    
    def __init__(self):
        self.compiled_log_patterns = [re.compile(p, re.IGNORECASE) for p in self.LOG_PATTERNS]
        self.loop_pattern = re.compile('|'.join(self.LOOP_PATTERNS))
        self.error_pattern = re.compile('|'.join(self.ERROR_HANDLER_PATTERNS))
        self.func_pattern = re.compile(self.FUNCTION_PATTERN)
    
    def scan_file(self, file_path: str) -> List[LogStatement]:
        """Scan a single Go file for log statements."""
        logs = []
        
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()
                lines = content.split('\n')
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            return logs
        
        current_function = "unknown"
        brace_depth = 0
        in_loop_depth = 0
        in_error_handler = False
        error_handler_brace_depth = 0
        loop_start_depths = []
        
        for line_num, line in enumerate(lines, 1):
            # Track function scope
            func_match = self.func_pattern.search(line)
            if func_match:
                current_function = func_match.group(1)
            
            # Track brace depth
            brace_depth += line.count('{') - line.count('}')
            
            # Track loop scope
            if self.loop_pattern.search(line):
                loop_start_depths.append(brace_depth)
            
            # Remove closed loops
            loop_start_depths = [d for d in loop_start_depths if d <= brace_depth]
            in_loop = len(loop_start_depths) > 0
            
            # Track error handler scope
            if self.error_pattern.search(line):
                in_error_handler = True
                error_handler_brace_depth = brace_depth
            
            if in_error_handler and brace_depth < error_handler_brace_depth:
                in_error_handler = False
            
            # Check for log statements
            for pattern in self.compiled_log_patterns:
                match = pattern.search(line)
                if match:
                    level = self._extract_level(match.group())
                    message = self._extract_message(line)

                    log = LogStatement(
                        file=file_path,
                        line=line_num,
                        column=match.start() + 1,
                        level=level,
                        message=message,
                        function=current_function,
                        statement=line.strip(),
                        in_loop=in_loop,
                        in_error_handler=in_error_handler,
                        language='go',
                        estimated_size_bytes=self._estimate_size(line),
                        message_template=extract_message_template(message)
                    )
                    logs.append(log)
                    break  # Only count one log per line
        
        return logs
    
    def _extract_level(self, match_text: str) -> str:
        """Extract log level from matched text."""
        match_lower = match_text.lower()
        
        if 'fatal' in match_lower:
            return 'FATAL'
        elif 'panic' in match_lower:
            return 'PANIC'
        elif 'error' in match_lower:
            return 'ERROR'
        elif 'warn' in match_lower:
            return 'WARN'
        elif 'info' in match_lower:
            return 'INFO'
        elif 'debug' in match_lower:
            return 'DEBUG'
        else:
            return 'INFO'  # Default
    
    def _extract_message(self, line: str) -> str:
        """Extract log message from line."""
        # Try to find string literal
        string_match = re.search(r'"([^"]*)"', line)
        if string_match:
            return string_match.group(1)
        return ""
    
    def _estimate_size(self, line: str) -> int:
        """Estimate log output size in bytes."""
        # Base overhead: timestamp, level, caller info
        base_overhead = 200
        
        # Message size
        message_match = re.search(r'"([^"]*)"', line)
        message_size = len(message_match.group(1)) if message_match else 50
        
        # Count structured fields (key, value pairs)
        field_count = line.count('",') + line.count('", ')
        field_overhead = field_count * 50
        
        return base_overhead + message_size + field_overhead


class PHPLogScanner:
    """Scans PHP source files for log statements."""

    LOG_PATTERNS = [
        # Laravel/Monolog patterns
        r'Log::(debug|info|notice|warning|error|critical|alert|emergency)\s*\(',
        r'\$this->log->(debug|info|notice|warning|error|critical|alert|emergency)\s*\(',
        r'logger\(\)->(debug|info|notice|warning|error|critical|alert|emergency)\s*\(',
        # Razorpay trace patterns
        r'\$this->trace->(debug|info|warning|error|critical)\s*\(',
        r'\$trace->(debug|info|warning|error|critical)\s*\(',
        r'app\s*\(\s*[\'"]trace[\'"]\s*\)->(debug|info|warning|error|critical)\s*\(',
        # Generic logger patterns
        r'\$logger->(debug|info|warning|error|critical)\s*\(',
        r'\$this->logger->(debug|info|warning|error|critical)\s*\(',
        # error_log function
        r'error_log\s*\(',
    ]

    LOOP_PATTERNS = [
        r'\bfor\s*\(',
        r'\bforeach\s*\(',
        r'\bwhile\s*\(',
    ]

    ERROR_HANDLER_PATTERNS = [
        r'catch\s*\(',
        r'if\s*\(\s*\$e\s*',
        r'if\s*\(\s*\$exception',
    ]

    def __init__(self):
        self.compiled_patterns = [re.compile(p, re.IGNORECASE) for p in self.LOG_PATTERNS]
        self.loop_pattern = re.compile('|'.join(self.LOOP_PATTERNS))
        self.error_pattern = re.compile('|'.join(self.ERROR_HANDLER_PATTERNS))

    def scan_file(self, file_path: str) -> List[LogStatement]:
        """Scan a single PHP file for log statements."""
        logs = []

        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()
                lines = content.split('\n')
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            return logs

        current_function = "unknown"
        brace_depth = 0
        in_loop_depth = 0
        in_error_handler = False
        error_handler_brace_depth = 0
        loop_start_depths = []

        func_pattern = re.compile(r'function\s+(\w+)\s*\(')

        for line_num, line in enumerate(lines, 1):
            # Track function scope
            func_match = func_pattern.search(line)
            if func_match:
                current_function = func_match.group(1)

            # Track brace depth
            brace_depth += line.count('{') - line.count('}')

            # Track loop scope
            if self.loop_pattern.search(line):
                loop_start_depths.append(brace_depth)

            loop_start_depths = [d for d in loop_start_depths if d <= brace_depth]
            in_loop = len(loop_start_depths) > 0

            # Track error handler scope
            if self.error_pattern.search(line):
                in_error_handler = True
                error_handler_brace_depth = brace_depth

            if in_error_handler and brace_depth < error_handler_brace_depth:
                in_error_handler = False

            # Check for log statements
            for pattern in self.compiled_patterns:
                match = pattern.search(line)
                if match:
                    level = self._extract_level(match.group())
                    message = self._extract_message(line)

                    log = LogStatement(
                        file=file_path,
                        line=line_num,
                        column=match.start() + 1,
                        level=level,
                        message=message,
                        function=current_function,
                        statement=line.strip(),
                        in_loop=in_loop,
                        in_error_handler=in_error_handler,
                        language='php',
                        estimated_size_bytes=self._estimate_size(line),
                        message_template=extract_message_template(message)
                    )
                    logs.append(log)
                    break

        return logs

    def _extract_level(self, match_text: str) -> str:
        """Extract log level from matched text."""
        match_lower = match_text.lower()

        if 'emergency' in match_lower or 'critical' in match_lower or 'alert' in match_lower:
            return 'FATAL'
        elif 'error' in match_lower:
            return 'ERROR'
        elif 'warning' in match_lower or 'notice' in match_lower:
            return 'WARN'
        elif 'info' in match_lower:
            return 'INFO'
        elif 'debug' in match_lower:
            return 'DEBUG'
        else:
            return 'INFO'

    def _extract_message(self, line: str) -> str:
        """Extract log message from line."""
        for quote in ['"', "'"]:
            match = re.search(f'{quote}([^{quote}]*){quote}', line)
            if match:
                return match.group(1)
        return ""

    def _estimate_size(self, line: str) -> int:
        """Estimate log output size in bytes."""
        base_overhead = 200
        message_match = re.search(r'["\']([^"\']*)["\']', line)
        message_size = len(message_match.group(1)) if message_match else 50
        return base_overhead + message_size


class TypeScriptLogScanner:
    """Scans TypeScript/JavaScript source files for log statements."""

    LOG_PATTERNS = [
        # Console methods
        r'console\.(log|debug|info|warn|error)\s*\(',
        # Winston/Bunyan/Pino patterns
        r'logger\.(debug|info|warn|warning|error|fatal)\s*\(',
        r'log\.(debug|info|warn|warning|error|fatal)\s*\(',
        r'this\.logger\.(debug|info|warn|warning|error|fatal)\s*\(',
        # Custom logging
        r'Logger\.(debug|info|warn|warning|error|fatal)\s*\(',
    ]

    LOOP_PATTERNS = [
        r'\bfor\s*\(',
        r'\bwhile\s*\(',
        r'\.forEach\s*\(',
        r'\.map\s*\(',
        r'\.filter\s*\(',
        r'for\s*\.\.\.\s*of',
        r'for\s*\.\.\.\s*in',
    ]

    ERROR_HANDLER_PATTERNS = [
        r'catch\s*\(',
        r'\.catch\s*\(',
        r'if\s*\(\s*error',
        r'if\s*\(\s*err\s*\)',
    ]

    def __init__(self):
        self.compiled_patterns = [re.compile(p, re.IGNORECASE) for p in self.LOG_PATTERNS]
        self.loop_pattern = re.compile('|'.join(self.LOOP_PATTERNS))
        self.error_pattern = re.compile('|'.join(self.ERROR_HANDLER_PATTERNS))

    def scan_file(self, file_path: str) -> List[LogStatement]:
        """Scan a single TypeScript/JavaScript file for log statements."""
        logs = []

        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                content = f.read()
                lines = content.split('\n')
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            return logs

        current_function = "unknown"
        brace_depth = 0
        in_error_handler = False
        error_handler_brace_depth = 0
        loop_start_depths = []

        # Match function, arrow function, method
        func_pattern = re.compile(r'(?:function\s+(\w+)|(\w+)\s*[=:]\s*(?:async\s*)?\(|(\w+)\s*\([^)]*\)\s*{)')

        for line_num, line in enumerate(lines, 1):
            # Track function scope
            func_match = func_pattern.search(line)
            if func_match:
                current_function = func_match.group(1) or func_match.group(2) or func_match.group(3) or "anonymous"

            # Track brace depth
            brace_depth += line.count('{') - line.count('}')

            # Track loop scope
            if self.loop_pattern.search(line):
                loop_start_depths.append(brace_depth)

            loop_start_depths = [d for d in loop_start_depths if d <= brace_depth]
            in_loop = len(loop_start_depths) > 0

            # Track error handler scope
            if self.error_pattern.search(line):
                in_error_handler = True
                error_handler_brace_depth = brace_depth

            if in_error_handler and brace_depth < error_handler_brace_depth:
                in_error_handler = False

            # Check for log statements
            for pattern in self.compiled_patterns:
                match = pattern.search(line)
                if match:
                    level = self._extract_level(match.group())
                    message = self._extract_message(line)

                    log = LogStatement(
                        file=file_path,
                        line=line_num,
                        column=match.start() + 1,
                        level=level,
                        message=message,
                        function=current_function,
                        statement=line.strip(),
                        in_loop=in_loop,
                        in_error_handler=in_error_handler,
                        language='typescript',
                        estimated_size_bytes=self._estimate_size(line),
                        message_template=extract_message_template(message)
                    )
                    logs.append(log)
                    break

        return logs

    def _extract_level(self, match_text: str) -> str:
        """Extract log level from matched text."""
        match_lower = match_text.lower()

        if 'fatal' in match_lower:
            return 'FATAL'
        elif 'error' in match_lower:
            return 'ERROR'
        elif 'warn' in match_lower:
            return 'WARN'
        elif 'info' in match_lower:
            return 'INFO'
        elif 'debug' in match_lower:
            return 'DEBUG'
        elif 'log' in match_lower:
            return 'INFO'  # console.log is typically INFO level
        else:
            return 'INFO'

    def _extract_message(self, line: str) -> str:
        """Extract log message from line."""
        # Try template literals first
        match = re.search(r'`([^`]*)`', line)
        if match:
            return match.group(1)
        # Then regular strings
        for quote in ['"', "'"]:
            match = re.search(f'{quote}([^{quote}]*){quote}', line)
            if match:
                return match.group(1)
        return ""

    def _estimate_size(self, line: str) -> int:
        """Estimate log output size in bytes."""
        base_overhead = 150
        message_match = re.search(r'["\'\`]([^"\'\`]*)["\'\`]', line)
        message_size = len(message_match.group(1)) if message_match else 50
        return base_overhead + message_size


class PythonLogScanner:
    """Scans Python source files for log statements."""

    LOG_PATTERNS = [
        r'logging\.(debug|info|warning|error|critical|exception)\s*\(',
        r'logger\.(debug|info|warning|error|critical|exception)\s*\(',
        r'log\.(debug|info|warning|error|critical|exception)\s*\(',
        r'self\.logger\.(debug|info|warning|error|critical|exception)\s*\(',
    ]
    
    def __init__(self):
        self.compiled_patterns = [re.compile(p, re.IGNORECASE) for p in self.LOG_PATTERNS]
    
    def scan_file(self, file_path: str) -> List[LogStatement]:
        """Scan a single Python file for log statements."""
        logs = []
        
        try:
            with open(file_path, 'r', encoding='utf-8', errors='ignore') as f:
                lines = f.readlines()
        except Exception as e:
            print(f"Error reading {file_path}: {e}")
            return logs
        
        current_function = "module"
        in_loop = False
        in_try_except = False
        indent_stack = []
        
        for line_num, line in enumerate(lines, 1):
            stripped = line.strip()
            indent = len(line) - len(line.lstrip())
            
            # Track function scope
            if stripped.startswith('def '):
                func_match = re.search(r'def\s+(\w+)\s*\(', stripped)
                if func_match:
                    current_function = func_match.group(1)
            
            # Track loop scope (simplified)
            if stripped.startswith(('for ', 'while ')):
                in_loop = True
                indent_stack.append(('loop', indent))
            
            # Track try/except scope
            if stripped.startswith('except'):
                in_try_except = True
                indent_stack.append(('except', indent))
            
            # Pop from stack when dedenting
            while indent_stack and indent <= indent_stack[-1][1]:
                scope_type, _ = indent_stack.pop()
                if scope_type == 'loop':
                    in_loop = any(s[0] == 'loop' for s in indent_stack)
                elif scope_type == 'except':
                    in_try_except = any(s[0] == 'except' for s in indent_stack)
            
            # Check for log statements
            for pattern in self.compiled_patterns:
                match = pattern.search(line)
                if match:
                    level = match.group(1).upper()
                    if level == 'EXCEPTION':
                        level = 'ERROR'
                    elif level == 'CRITICAL':
                        level = 'FATAL'

                    message = self._extract_message(line)
                    log = LogStatement(
                        file=file_path,
                        line=line_num,
                        column=match.start() + 1,
                        level=level,
                        message=message,
                        function=current_function,
                        statement=stripped,
                        in_loop=in_loop,
                        in_error_handler=in_try_except,
                        language='python',
                        estimated_size_bytes=self._estimate_size(line),
                        message_template=extract_message_template(message)
                    )
                    logs.append(log)
                    break
        
        return logs
    
    def _extract_message(self, line: str) -> str:
        """Extract log message from line."""
        for quote in ['"', "'"]:
            match = re.search(f'{quote}([^{quote}]*){quote}', line)
            if match:
                return match.group(1)
        return ""
    
    def _estimate_size(self, line: str) -> int:
        """Estimate log output size in bytes."""
        base_overhead = 150
        message_match = re.search(r'["\']([^"\']*)["\']', line)
        message_size = len(message_match.group(1)) if message_match else 50
        return base_overhead + message_size


def detect_language(repo_path: str) -> str:
    """Auto-detect primary language of a repository."""
    repo_path = Path(repo_path)

    # Count files by extension
    counts = {
        'go': 0,
        'php': 0,
        'typescript': 0,
        'python': 0,
    }

    for ext, lang in [('.go', 'go'), ('.php', 'php'), ('.ts', 'typescript'),
                       ('.tsx', 'typescript'), ('.js', 'typescript'), ('.py', 'python')]:
        counts[lang] += len(list(repo_path.glob(f'**/*{ext}')))

    # Return language with most files
    if max(counts.values()) == 0:
        return 'go'  # Default
    return max(counts, key=counts.get)


def get_scanner_for_language(language: str):
    """Get the appropriate scanner for a language."""
    scanners = {
        'go': GoLogScanner,
        'php': PHPLogScanner,
        'typescript': TypeScriptLogScanner,
        'javascript': TypeScriptLogScanner,
        'python': PythonLogScanner,
    }
    scanner_class = scanners.get(language.lower())
    if scanner_class:
        return scanner_class()
    return None


def get_file_patterns(language: str) -> tuple:
    """Get include and exclude patterns for a language."""
    # Common exclusions for all languages (from Rajeev's approach)
    common_excludes = [f'**/{d}/**' for d in EXCLUDED_DIRS]

    patterns = {
        'go': {
            'include': ['**/*.go'],
            'exclude': common_excludes + ['**/*_test.go', '**/*_mock.go']
        },
        'php': {
            'include': ['**/*.php'],
            'exclude': common_excludes + ['**/*Test.php', '**/*_test.php']
        },
        'typescript': {
            'include': ['**/*.ts', '**/*.tsx', '**/*.js', '**/*.jsx'],
            'exclude': common_excludes + ['**/*.test.ts', '**/*.spec.ts', '**/*.test.js', '**/*.spec.js', '**/*.d.ts']
        },
        'python': {
            'include': ['**/*.py'],
            'exclude': common_excludes + ['**/test_*.py', '**/*_test.py', '**/conftest.py']
        }
    }
    lang_patterns = patterns.get(language.lower(), patterns['go'])
    return lang_patterns['include'], lang_patterns['exclude']


def scan_repository(
    repo_path: str,
    language: str = 'auto',
    include_patterns: List[str] = None,
    exclude_patterns: List[str] = None
) -> List[LogStatement]:
    """
    Scan an entire repository for log statements.

    Args:
        repo_path: Path to repository root
        language: Programming language (go, php, typescript, python, or 'auto' for detection)
        include_patterns: Glob patterns to include
        exclude_patterns: Glob patterns to exclude

    Returns:
        List of detected log statements
    """
    repo_path_obj = Path(repo_path)

    # Auto-detect language if needed
    if language == 'auto':
        language = detect_language(repo_path)
        print(f"  Auto-detected language: {language}")

    # Get scanner
    scanner = get_scanner_for_language(language)
    if not scanner:
        print(f"  Warning: No scanner for language '{language}', trying Go scanner")
        scanner = GoLogScanner()
        language = 'go'

    # Get patterns
    if include_patterns is None or exclude_patterns is None:
        default_include, default_exclude = get_file_patterns(language)
        include_patterns = include_patterns or default_include
        exclude_patterns = exclude_patterns or default_exclude

    all_logs = []

    # Find all matching files
    for pattern in include_patterns:
        for file_path in repo_path_obj.glob(pattern):
            # Check exclusions
            rel_path = str(file_path.relative_to(repo_path_obj))
            excluded = False
            for exc in exclude_patterns:
                if Path(rel_path).match(exc.replace('**/', '')):
                    excluded = True
                    break

            if not excluded and file_path.is_file():
                logs = scanner.scan_file(str(file_path))
                all_logs.extend(logs)

    return all_logs


def scan_repository_all_languages(repo_path: str) -> List[LogStatement]:
    """
    Scan a repository for log statements in ALL supported languages.

    This scans for Go, PHP, TypeScript, and Python logs in a single pass.

    Args:
        repo_path: Path to repository root

    Returns:
        List of detected log statements from all languages
    """
    all_logs = []

    for language in ['go', 'php', 'typescript', 'python']:
        scanner = get_scanner_for_language(language)
        include_patterns, exclude_patterns = get_file_patterns(language)

        repo_path_obj = Path(repo_path)

        for pattern in include_patterns:
            for file_path in repo_path_obj.glob(pattern):
                rel_path = str(file_path.relative_to(repo_path_obj))
                excluded = False
                for exc in exclude_patterns:
                    if Path(rel_path).match(exc.replace('**/', '')):
                        excluded = True
                        break

                if not excluded and file_path.is_file():
                    logs = scanner.scan_file(str(file_path))
                    all_logs.extend(logs)

    return all_logs


def generate_summary(logs: List[LogStatement]) -> Dict:
    """Generate a summary of scanned logs."""
    summary = {
        'total_logs': len(logs),
        'by_level': {},
        'by_file': {},
        'in_loops': 0,
        'in_error_handlers': 0,
        'estimated_daily_bytes': 0,
    }
    
    for log in logs:
        # Count by level
        level = log.level
        summary['by_level'][level] = summary['by_level'].get(level, 0) + 1
        
        # Count by file
        file = log.file
        summary['by_file'][file] = summary['by_file'].get(file, 0) + 1
        
        # Count context flags
        if log.in_loop:
            summary['in_loops'] += 1
        if log.in_error_handler:
            summary['in_error_handlers'] += 1
        
        # Sum estimated size
        summary['estimated_daily_bytes'] += log.estimated_size_bytes
    
    return summary


def main():
    parser = argparse.ArgumentParser(description='Scan repository for log statements')
    parser.add_argument('repo_path', help='Path to repository')
    parser.add_argument('--language', '-l', default='auto',
                        choices=['auto', 'go', 'php', 'typescript', 'javascript', 'python'],
                        help='Programming language (default: auto-detect)')
    parser.add_argument('--all-languages', '-a', action='store_true',
                        help='Scan for all supported languages')
    parser.add_argument('--output', '-o', default='logs.json',
                        help='Output file path')
    parser.add_argument('--summary', '-s', action='store_true',
                        help='Print summary to stdout')

    args = parser.parse_args()

    print(f"Scanning {args.repo_path}...")

    if args.all_languages:
        print("Scanning for ALL languages (Go, PHP, TypeScript, Python)...")
        logs = scan_repository_all_languages(args.repo_path)
    else:
        logs = scan_repository(args.repo_path, args.language)
    
    print(f"Found {len(logs)} log statements")
    
    # Generate summary
    summary = generate_summary(logs)
    
    if args.summary:
        print("\n=== Summary ===")
        print(f"Total logs: {summary['total_logs']}")
        print(f"By level: {summary['by_level']}")
        print(f"In loops: {summary['in_loops']}")
        print(f"In error handlers: {summary['in_error_handlers']}")
    
    # Write output
    output = {
        'summary': summary,
        'logs': [log.to_dict() for log in logs]
    }
    
    with open(args.output, 'w') as f:
        json.dump(output, f, indent=2)
    
    print(f"Results written to {args.output}")


if __name__ == '__main__':
    main()
