#!/usr/bin/env python3
"""
Log Volume Analyzer for pg-router
Analyzes log statements and estimates their volume impact
"""

import re
import json
from pathlib import Path
from collections import defaultdict
from dataclasses import dataclass, asdict
from typing import Dict, List, Optional

@dataclass
class LogStatement:
    file_path: str
    line_number: int
    level: str
    message: str
    function: str = ""
    in_loop: bool = False
    in_error_handler: bool = False
    estimated_rps: float = 0.0
    estimated_daily_logs: int = 0
    estimated_daily_bytes: int = 0
    estimated_daily_units: float = 0.0

class LogAnalyzer:
    def __init__(self, repo_path: str, config: dict):
        self.repo_path = Path(repo_path)
        self.config = config
        self.log_statements: List[LogStatement] = []
        self.stats = {
            "total": 0,
            "by_level": defaultdict(int),
            "by_directory": defaultdict(int),
            "in_loops": 0,
            "in_error_handlers": 0,
        }

    def scan_repository(self):
        """Scan repository for log statements"""
        # Pattern to match log statements
        log_pattern = re.compile(
            r'(?:logger\.Log\(ctx\)\.|lgr\.Logger\(ctx\)\.)' +
            r'(Info|Infow|Error|Errorw|Debug|Debugw|Warn|Warnw|Fatal|Fatalw)'
        )

        # Search all .go files
        go_files = list(self.repo_path.rglob("*.go"))

        for go_file in go_files:
            # Skip test files and vendor
            if "_test.go" in go_file.name or "vendor" in go_file.parts:
                continue

            try:
                with open(go_file, 'r', encoding='utf-8', errors='ignore') as f:
                    lines = f.readlines()

                for line_num, line in enumerate(lines, 1):
                    match = log_pattern.search(line)
                    if match:
                        level = match.group(1).replace('w', '').upper()

                        # Extract message (simplified)
                        message = line.strip()

                        # Detect context
                        in_loop = self._check_in_loop(lines, line_num)
                        in_error = self._check_in_error_handler(lines, line_num)

                        # Determine function
                        function = self._find_function_name(lines, line_num)

                        # Create log statement
                        log_stmt = LogStatement(
                            file_path=str(go_file.relative_to(self.repo_path)),
                            line_number=line_num,
                            level=level,
                            message=message[:200],  # Truncate long messages
                            function=function,
                            in_loop=in_loop,
                            in_error_handler=in_error
                        )

                        self.log_statements.append(log_stmt)
                        self.stats["total"] += 1
                        self.stats["by_level"][level] += 1

                        if in_loop:
                            self.stats["in_loops"] += 1
                        if in_error:
                            self.stats["in_error_handlers"] += 1

            except Exception as e:
                print(f"Error reading {go_file}: {e}")

    def _check_in_loop(self, lines: List[str], current_line: int) -> bool:
        """Check if log statement is inside a loop"""
        # Look back up to 20 lines
        start = max(0, current_line - 20)
        context = ''.join(lines[start:current_line])

        # Simple heuristic: look for for loops
        return bool(re.search(r'\bfor\s+', context))

    def _check_in_error_handler(self, lines: List[str], current_line: int) -> bool:
        """Check if log statement is in error handler"""
        # Look back up to 10 lines
        start = max(0, current_line - 10)
        context = ''.join(lines[start:current_line])

        # Look for if err != nil or error handling
        return bool(re.search(r'if\s+.*err\s*!=\s*nil', context))

    def _find_function_name(self, lines: List[str], current_line: int) -> str:
        """Find the function name containing this log statement"""
        # Look backwards for function definition
        for i in range(current_line - 1, max(0, current_line - 50), -1):
            match = re.match(r'func\s+(?:\([^)]+\)\s+)?(\w+)', lines[i])
            if match:
                return match.group(1)
        return "unknown"

    def estimate_volume(self):
        """Estimate log volume for each statement"""
        avg_rps = self.config['traffic']['avg_rps']
        default_error_rate = self.config['analysis']['default_error_rate']
        base_log_overhead = self.config['analysis']['base_log_overhead']
        field_overhead = self.config['analysis']['field_overhead']

        for log_stmt in self.log_statements:
            # Estimate trigger rate based on context
            if log_stmt.level == "FATAL":
                # Fatal logs only on startup/critical errors
                trigger_rate = 0.0001
            elif log_stmt.level == "ERROR":
                # Errors based on error rate
                if log_stmt.in_error_handler:
                    trigger_rate = default_error_rate
                else:
                    trigger_rate = 0.001
            elif log_stmt.level == "WARN":
                trigger_rate = 0.01
            elif log_stmt.level == "INFO":
                # Info logs trigger more frequently
                if log_stmt.in_loop:
                    trigger_rate = 10.0  # 10x multiplier for loops
                else:
                    trigger_rate = 1.0
            elif log_stmt.level == "DEBUG":
                # Debug should be disabled in prod, but let's count it
                if log_stmt.in_loop:
                    trigger_rate = 10.0
                else:
                    trigger_rate = 1.0
            else:
                trigger_rate = 1.0

            # Calculate RPS
            log_stmt.estimated_rps = avg_rps * trigger_rate

            # Daily logs = RPS * seconds_per_day
            log_stmt.estimated_daily_logs = int(log_stmt.estimated_rps * 86400)

            # Estimate log size (base + message + fields)
            avg_log_size = base_log_overhead + len(log_stmt.message) + (field_overhead * 3)
            log_stmt.estimated_daily_bytes = log_stmt.estimated_daily_logs * avg_log_size

            # Coralogix units (1 unit = 1 MB)
            log_stmt.estimated_daily_units = log_stmt.estimated_daily_bytes / 1_000_000

    def get_high_impact_logs(self, top_n: int = 50) -> List[LogStatement]:
        """Get top N highest impact log statements"""
        return sorted(
            self.log_statements,
            key=lambda x: x.estimated_daily_units,
            reverse=True
        )[:top_n]

    def generate_report(self) -> dict:
        """Generate analysis report"""
        total_daily_units = sum(log.estimated_daily_units for log in self.log_statements)
        total_daily_bytes = sum(log.estimated_daily_bytes for log in self.log_statements)

        report = {
            "summary": {
                "total_log_statements": self.stats["total"],
                "total_daily_logs": sum(log.estimated_daily_logs for log in self.log_statements),
                "total_daily_bytes": total_daily_bytes,
                "total_daily_gb": total_daily_bytes / 1_000_000_000,
                "total_daily_units": total_daily_units,
                "assigned_quota": self.config['quota']['assigned_units'],
                "quota_utilization_pct": (total_daily_units / self.config['quota']['assigned_units']) * 100
            },
            "by_level": dict(self.stats["by_level"]),
            "high_impact_logs": [
                {
                    "file": log.file_path,
                    "line": log.line_number,
                    "level": log.level,
                    "function": log.function,
                    "daily_units": round(log.estimated_daily_units, 2),
                    "daily_logs": log.estimated_daily_logs,
                    "in_loop": log.in_loop,
                    "in_error_handler": log.in_error_handler,
                    "message_preview": log.message[:100]
                }
                for log in self.get_high_impact_logs(50)
            ],
            "recommendations": self.generate_recommendations()
        }

        return report

    def generate_recommendations(self) -> List[dict]:
        """Generate optimization recommendations"""
        recommendations = []

        # Find INFO logs in loops
        info_in_loops = [log for log in self.log_statements
                         if log.level == "INFO" and log.in_loop]
        if info_in_loops:
            savings = sum(log.estimated_daily_units for log in info_in_loops)
            recommendations.append({
                "priority": "HIGH",
                "category": "Remove logs in loops",
                "count": len(info_in_loops),
                "estimated_savings_units": round(savings, 2),
                "description": f"Found {len(info_in_loops)} INFO logs inside loops. These should be removed or moved outside loops."
            })

        # Find DEBUG logs
        debug_logs = [log for log in self.log_statements if log.level == "DEBUG"]
        if debug_logs:
            savings = sum(log.estimated_daily_units for log in debug_logs)
            recommendations.append({
                "priority": "MEDIUM",
                "category": "DEBUG logs in production",
                "count": len(debug_logs),
                "estimated_savings_units": round(savings, 2),
                "description": f"Found {len(debug_logs)} DEBUG logs. Ensure these are disabled in production."
            })

        # Find high-frequency INFO logs
        high_freq_info = [log for log in self.log_statements
                          if log.level == "INFO" and log.estimated_daily_units > 10]
        if high_freq_info:
            savings = sum(log.estimated_daily_units * 0.5 for log in high_freq_info)  # 50% reduction via sampling
            recommendations.append({
                "priority": "HIGH",
                "category": "Add sampling to high-frequency logs",
                "count": len(high_freq_info),
                "estimated_savings_units": round(savings, 2),
                "description": f"Found {len(high_freq_info)} high-frequency INFO logs. Consider adding sampling (1% sample rate)."
            })

        return recommendations

def main():
    import sys

    if len(sys.argv) < 2:
        print("Usage: python analyze_logs.py <repo_path>")
        sys.exit(1)

    repo_path = sys.argv[1]

    # Load config
    config = {
        "quota": {"assigned_units": 480},
        "traffic": {"avg_rps": 500},
        "analysis": {
            "default_error_rate": 0.01,
            "base_log_overhead": 200,
            "field_overhead": 50
        }
    }

    analyzer = LogAnalyzer(repo_path, config)
    print("Scanning repository...")
    analyzer.scan_repository()

    print("Estimating volume...")
    analyzer.estimate_volume()

    print("Generating report...")
    report = analyzer.generate_report()

    # Print report
    print(json.dumps(report, indent=2))

if __name__ == "__main__":
    main()
