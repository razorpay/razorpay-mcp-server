#!/usr/bin/env python3
"""
Coralogix Integration for Log Volume Optimizer

Handles loading, parsing, and matching real log frequency data from Coralogix
to code-scanned log statements for accurate decision-making.

Supports multiple data sources:
- Coralogix MCP query results (JSON)
- Exported CSV from Coralogix UI
- Manual frequency data

The AI agent (Claude Code) queries Coralogix MCP and saves results,
then this module loads and matches them to code-scanned logs.

Usage:
    # As a library
    from coralogix_integration import CoralogixData
    cx = CoralogixData()
    cx.load_from_json("reports/pg-router-coralogix.json")
    freq = cx.get_daily_frequency("Processing payment for merchant")

    # CLI
    python coralogix_integration.py --input data.json --summary
    python coralogix_integration.py --input data.csv --top 20
    python coralogix_integration.py --generate-query --app pg-router
"""

import csv
import json
import os
import re
from dataclasses import dataclass, field
from datetime import datetime
from difflib import SequenceMatcher
from typing import Dict, List, Optional, Tuple

# Import the template extractor from scan_logs
try:
    from scan_logs import extract_message_template
except ImportError:
    import sys
    sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
    from scan_logs import extract_message_template


@dataclass
class FrequencyEntry:
    """A single log frequency entry from Coralogix."""
    raw_message: str
    template: str
    count: int              # Count in the query time window
    daily_count: int        # Extrapolated daily count
    daily_gb: float = 0.0   # Estimated daily GB


@dataclass
class MatchResult:
    """Result of matching a code log to Coralogix data."""
    code_file: str
    code_line: int
    code_message: str
    code_template: str
    coralogix_message: str
    coralogix_template: str
    match_score: float      # 0.0 to 1.0
    daily_frequency: int
    match_method: str       # 'exact', 'template', 'substring', 'fuzzy', 'none'


class CoralogixData:
    """
    Handles Coralogix log frequency data for the optimizer.

    Workflow:
    1. AI agent queries Coralogix MCP:
       mcp_coralogix-server_get_logs with DataPrime query
    2. AI saves results to JSON file
    3. This class loads and indexes the data
    4. Provides frequency lookup for the decision engine
    """

    # Default query window
    DEFAULT_QUERY_MINUTES = 15
    MINUTES_PER_DAY = 1440

    def __init__(self):
        self.entries: List[FrequencyEntry] = []
        self.template_index: Dict[str, FrequencyEntry] = {}  # template -> best entry
        self.application: str = ""
        self.query_minutes: int = self.DEFAULT_QUERY_MINUTES
        self._loaded = False

    @property
    def is_loaded(self) -> bool:
        return self._loaded and len(self.entries) > 0

    def load_from_json(self, json_path: str, query_minutes: int = None):
        """
        Load frequency data from a JSON file.

        Supports multiple formats:
        1. MCP response: {"results": [{"$d.message": "...", "_count": N}, ...]}
        2. Skill format: {"application": "...", "query_minutes": 15, "frequencies": [...]}
        3. Raw list: [{"message": "...", "count": N}, ...]
        4. Flat dict: {"message1": count1, "message2": count2}
        """
        with open(json_path, 'r') as f:
            data = json.load(f)

        if query_minutes:
            self.query_minutes = query_minutes

        if isinstance(data, dict):
            self.application = data.get('application', '')
            if 'query_minutes' in data:
                self.query_minutes = data['query_minutes']

            # MCP response format
            if 'results' in data:
                self._parse_entries(data['results'])
            # Skill format
            elif 'frequencies' in data:
                self._parse_entries(data['frequencies'])
            # Flat dict {message: count}
            else:
                for msg, count in data.items():
                    if isinstance(count, (int, float)) and msg not in (
                        'application', 'query_minutes', 'timestamp', 'query'
                    ):
                        self._add_entry(msg, int(count))
        elif isinstance(data, list):
            self._parse_entries(data)

        self._build_index()
        self._loaded = True
        print(f"  Loaded {len(self.entries)} Coralogix frequency entries from {json_path}")
        if self.entries:
            top = self.entries[0] if self.entries else None
            print(f"  Top message: [{top.daily_count:,}/day] {top.raw_message[:60]}...")

    def load_from_csv(self, csv_path: str, query_minutes: int = None):
        """
        Load frequency data from CSV export (Monark's approach).

        Expected columns: message (or $d.message), count (or _count)
        """
        if query_minutes:
            self.query_minutes = query_minutes

        with open(csv_path, 'r', encoding='utf-8', errors='ignore') as f:
            reader = csv.DictReader(f)
            for row in reader:
                # Try different column name conventions
                message = (row.get('message') or row.get('$d.message') or
                          row.get('msg') or row.get('log_message') or '')
                count_str = (row.get('count') or row.get('_count') or
                           row.get('Count') or row.get('frequency') or '0')

                try:
                    count = int(float(count_str))
                except (ValueError, TypeError):
                    count = 0

                if message and count > 0:
                    self._add_entry(message, count)

        self._build_index()
        self._loaded = True
        print(f"  Loaded {len(self.entries)} Coralogix frequency entries from {csv_path}")

    def load_from_mcp_response(self, response_data, query_minutes: int = None):
        """
        Load directly from MCP tool response data.

        The AI agent can pass the raw MCP response (dict, list, or JSON string).
        """
        if query_minutes:
            self.query_minutes = query_minutes

        if isinstance(response_data, str):
            try:
                response_data = json.loads(response_data)
            except json.JSONDecodeError:
                # Try line-separated JSON
                for line in response_data.strip().split('\n'):
                    try:
                        entry = json.loads(line)
                        msg = entry.get('$d.message') or entry.get('message') or ''
                        count = entry.get('_count') or entry.get('count') or 0
                        if msg and count:
                            self._add_entry(str(msg), int(count))
                    except (json.JSONDecodeError, ValueError):
                        continue
                self._build_index()
                self._loaded = True
                return

        if isinstance(response_data, list):
            self._parse_entries(response_data)
        elif isinstance(response_data, dict):
            if 'results' in response_data:
                self._parse_entries(response_data['results'])
            elif 'frequencies' in response_data:
                self._parse_entries(response_data['frequencies'])

        self._build_index()
        self._loaded = True
        print(f"  Loaded {len(self.entries)} frequency entries from MCP response")

    def _parse_entries(self, entries: list):
        """Parse a list of entry dicts."""
        for item in entries:
            if isinstance(item, dict):
                # Try all known key names for message and count
                msg = (item.get('$d.message') or item.get('message') or
                      item.get('$d.msg') or item.get('msg') or
                      item.get('log_message') or '')
                count = (item.get('_count') or item.get('count') or
                        item.get('frequency') or item.get('Count') or 0)

                if isinstance(count, str):
                    try:
                        count = int(float(count))
                    except (ValueError, TypeError):
                        count = 0

                if msg and count > 0:
                    self._add_entry(str(msg), int(count))

    def _add_entry(self, raw_message: str, count: int):
        """Add a frequency entry with template extraction and daily extrapolation."""
        template = extract_message_template(raw_message)
        extrapolation_factor = self.MINUTES_PER_DAY / self.query_minutes
        daily_count = int(count * extrapolation_factor)

        # Estimate daily GB (use actual message size or 300 bytes default)
        avg_log_size = max(len(raw_message.encode('utf-8', errors='ignore')), 200)
        daily_gb = (daily_count * avg_log_size) / (1024 ** 3)

        entry = FrequencyEntry(
            raw_message=raw_message,
            template=template,
            count=count,
            daily_count=daily_count,
            daily_gb=daily_gb
        )
        self.entries.append(entry)

    def _build_index(self):
        """Build template index for fast lookups. Higher frequency wins for duplicates."""
        self.template_index = {}
        for entry in self.entries:
            if not entry.template:
                continue
            existing = self.template_index.get(entry.template)
            if existing is None or entry.daily_count > existing.daily_count:
                self.template_index[entry.template] = entry

        # Sort entries by daily_count descending for top_offenders
        self.entries.sort(key=lambda e: e.daily_count, reverse=True)

    def get_daily_frequency(self, message: str) -> int:
        """
        Get the estimated daily frequency for a log message.

        Matching strategy (in order):
        1. Exact template match
        2. Substring/contains match (code template is often shorter)
        3. Fuzzy match (>80% similarity)

        Returns 0 if no match found.
        """
        if not self._loaded:
            return 0

        freq, _, _ = self.get_match_details(message)
        return freq

    def get_match_details(self, message: str) -> Tuple[int, str, float]:
        """
        Get frequency with match method details.

        Returns: (daily_frequency, match_method, match_score)
        """
        if not self._loaded or not message:
            return 0, 'none', 0.0

        template = extract_message_template(message)
        if not template:
            return 0, 'none', 0.0

        # 1. Exact template match
        entry = self.template_index.get(template)
        if entry:
            return entry.daily_count, 'exact', 1.0

        # 2. Substring match (code template is often a prefix of actual log message)
        best_substring = None
        best_substring_len = 0
        for t, e in self.template_index.items():
            if len(template) >= 10:  # Only substring match if template is meaningful
                if template in t and len(template) > best_substring_len:
                    best_substring = e
                    best_substring_len = len(template)
                elif t in template and len(t) > best_substring_len:
                    best_substring = e
                    best_substring_len = len(t)

        if best_substring and best_substring_len >= 10:
            return best_substring.daily_count, 'substring', 0.9

        # 3. Fuzzy match (expensive - only for remaining)
        best_score = 0.0
        best_entry = None
        template_lower = template.lower()
        for t, e in self.template_index.items():
            if t:
                score = SequenceMatcher(None, template_lower, t.lower()).ratio()
                if score > best_score and score >= 0.75:
                    best_score = score
                    best_entry = e

        if best_entry:
            return best_entry.daily_count, 'fuzzy', best_score

        return 0, 'none', 0.0

    def match_to_code_logs(self, scanned_logs: List[Dict]) -> Tuple[List[MatchResult], Dict]:
        """
        Match all scanned code logs to Coralogix frequency data.

        Args:
            scanned_logs: List of dicts with 'file', 'line', 'message', 'message_template' keys

        Returns:
            Tuple of (match_results, summary_dict)
        """
        results = []
        stats = {'exact': 0, 'substring': 0, 'fuzzy': 0, 'none': 0}

        for log in scanned_logs:
            message = log.get('message', '') or log.get('message_template', '')
            freq, method, score = self.get_match_details(message)

            coralogix_msg = ''
            coralogix_template = ''
            if method != 'none':
                # Find the matched Coralogix entry for reporting
                code_template = extract_message_template(message)
                matched_entry = self.template_index.get(code_template)
                if matched_entry:
                    coralogix_msg = matched_entry.raw_message[:100]
                    coralogix_template = matched_entry.template

            stats[method] = stats.get(method, 0) + 1

            results.append(MatchResult(
                code_file=log.get('file', ''),
                code_line=log.get('line', 0),
                code_message=message[:100],
                code_template=extract_message_template(message),
                coralogix_message=coralogix_msg,
                coralogix_template=coralogix_template,
                match_score=score,
                daily_frequency=freq,
                match_method=method
            ))

        total = len(results)
        matched = total - stats.get('none', 0)
        print(f"  Coralogix matching: {matched}/{total} code logs matched "
              f"(exact={stats['exact']}, substring={stats['substring']}, "
              f"fuzzy={stats['fuzzy']}, unmatched={stats['none']})")

        summary = {
            'total_code_logs': total,
            'matched': matched,
            'unmatched': stats.get('none', 0),
            'match_rate': round(matched / total * 100, 1) if total > 0 else 0,
            'by_method': stats
        }

        return results, summary

    def build_frequency_map(self) -> Dict[str, int]:
        """
        Build a template -> daily_frequency map for use with DecisionEngine.

        Returns a dict that can be passed directly to decision_engine.batch_decide()
        or used in github_api_prs.py.
        """
        freq_map = {}
        for template, entry in self.template_index.items():
            freq_map[template] = entry.daily_count
        return freq_map

    def get_top_offenders(self, limit: int = 20) -> List[FrequencyEntry]:
        """Get the top N highest frequency log messages."""
        return self.entries[:limit]

    def get_summary(self) -> Dict:
        """Get a summary of the loaded Coralogix data."""
        if not self.entries:
            return {"loaded": False, "entries": 0}

        total_daily = sum(e.daily_count for e in self.entries)
        total_daily_gb = sum(e.daily_gb for e in self.entries)

        return {
            "loaded": True,
            "application": self.application,
            "entries": len(self.entries),
            "unique_templates": len(self.template_index),
            "query_window_minutes": self.query_minutes,
            "total_daily_logs": total_daily,
            "total_daily_gb": round(total_daily_gb, 2),
            "top_10": [
                {
                    "message": e.raw_message[:80],
                    "daily_count": e.daily_count,
                    "daily_gb": round(e.daily_gb, 4)
                }
                for e in self.get_top_offenders(10)
            ]
        }


# ─── Helper Functions for AI Agent ────────────────────────────────────────────

def generate_coralogix_query(application_name: str, subsystem: str = None,
                             time_minutes: int = 15, limit: int = 500) -> str:
    """
    Generate the DataPrime query to run via Coralogix MCP.

    The AI agent should call:
        mcp_coralogix-server_get_logs(
            query=<this output>,
            start_date=<now - time_minutes>,
            end_date=<now>,
            limit=<limit>
        )
    """
    query = f"source logs | filter $l.applicationname == '{application_name}'"
    if subsystem:
        query += f" | filter $l.subsystemname == '{subsystem}'"
    query += f" | countby $d.message | sortby _count desc | limit {limit}"
    return query


def generate_cost_query(application_name: str) -> str:
    """
    Generate the full cost analysis DataPrime query.

    This gives per-message cost estimation based on priority class.
    """
    return f"""source logs
| filter $l.applicationname == '{application_name}'
| create size_bytes from $d:string.length()+$l:string.length()+$m:string.length()
| create msg_string from substr(trim($d.message), 0, 500)
| groupby msg_string, $m.priorityclass, $m.timestamp/5m as per_5_min
    sum(size_bytes)/(1024*1024*1024) as bytes_gb_per_5_min
| groupby msg_string, priorityclass
    agg avg(bytes_gb_per_5_min) as avg_size
| create estimated_cost_month from case {{
    priorityclass == 'low' -> round(avg_size * 12 * 12 * 30 * 0.59 * 0.12),
    priorityclass == 'medium' -> round(avg_size * 12 * 12 * 30 * 0.59 * 0.32),
    priorityclass == 'high' -> round(avg_size * 12 * 12 * 30 * 0.59 * 0.75)
}}
| sortby estimated_cost_month desc"""


def generate_units_query(application_name: str) -> str:
    """Generate PromQL query for Coralogix units consumption."""
    return f'sum(cx_data_usage_units{{application_name="{application_name}"}}) by (application_name)'


def save_coralogix_results(results: list, application: str, output_path: str,
                           query_minutes: int = 15, query: str = ""):
    """
    Save MCP query results to a JSON file for later use.

    The AI agent should call this after getting MCP results:
    1. Call mcp_coralogix-server_get_logs(...)
    2. Parse the results
    3. Save with this function
    4. Then github_api_prs.py --coralogix-data <path> uses it
    """
    data = {
        "application": application,
        "query_minutes": query_minutes,
        "query": query,
        "timestamp": datetime.now().isoformat(),
        "entry_count": len(results),
        "frequencies": results
    }

    os.makedirs(os.path.dirname(output_path) or '.', exist_ok=True)
    with open(output_path, 'w') as f:
        json.dump(data, f, indent=2)
    print(f"  Saved {len(results)} frequency entries to {output_path}")


def create_frequency_file_from_mcp(mcp_results: list, application: str,
                                    output_dir: str = "reports",
                                    query_minutes: int = 15) -> str:
    """
    Convenience function: parse MCP results and save to standard location.

    Returns the path to the created file.
    """
    output_path = os.path.join(output_dir, f"{application}-coralogix-frequencies.json")

    # Parse MCP results into our format
    frequencies = []
    for item in mcp_results:
        if isinstance(item, dict):
            msg = (item.get('$d.message') or item.get('message') or
                  item.get('msg') or '')
            count = item.get('_count') or item.get('count') or 0
            if msg and count:
                frequencies.append({
                    "message": str(msg),
                    "count": int(count)
                })

    save_coralogix_results(frequencies, application, output_path, query_minutes)
    return output_path


# ─── CLI ──────────────────────────────────────────────────────────────────────

if __name__ == '__main__':
    import argparse

    parser = argparse.ArgumentParser(
        description='Coralogix data integration for Log Volume Optimizer'
    )
    subparsers = parser.add_subparsers(dest='command', help='Command')

    # Load and analyze command
    load_parser = subparsers.add_parser('analyze', help='Load and analyze Coralogix data')
    load_parser.add_argument('--input', '-i', required=True,
                            help='Input file (JSON or CSV)')
    load_parser.add_argument('--query-minutes', type=int, default=15,
                            help='Query time window in minutes (default: 15)')
    load_parser.add_argument('--application', '-a', default='',
                            help='Application name')
    load_parser.add_argument('--top', type=int, default=20,
                            help='Show top N offenders (default: 20)')
    load_parser.add_argument('--output', '-o', help='Save processed data to JSON')

    # Generate query command
    query_parser = subparsers.add_parser('query', help='Generate Coralogix queries')
    query_parser.add_argument('--app', '-a', required=True,
                             help='Application name')
    query_parser.add_argument('--subsystem', '-s', help='Subsystem filter')
    query_parser.add_argument('--type', choices=['frequency', 'cost', 'units'],
                             default='frequency', help='Query type')

    # Match command
    match_parser = subparsers.add_parser('match', help='Match code logs to Coralogix data')
    match_parser.add_argument('--coralogix', '-c', required=True,
                             help='Coralogix data file (JSON/CSV)')
    match_parser.add_argument('--scan-results', '-s', required=True,
                             help='Scan results from scan_logs.py (JSON)')
    match_parser.add_argument('--query-minutes', type=int, default=15,
                             help='Query time window')
    match_parser.add_argument('--output', '-o', help='Output matched results')

    args = parser.parse_args()

    if args.command == 'analyze':
        cx = CoralogixData()
        if args.input.endswith('.csv'):
            cx.load_from_csv(args.input, query_minutes=args.query_minutes)
        else:
            cx.load_from_json(args.input, query_minutes=args.query_minutes)

        if args.application:
            cx.application = args.application

        # Print summary
        summary = cx.get_summary()
        print(f"\n{'='*60}")
        print(f"CORALOGIX DATA SUMMARY")
        print(f"{'='*60}")
        print(f"Application: {summary.get('application', 'unknown')}")
        print(f"Entries: {summary['entries']}")
        print(f"Unique templates: {summary['unique_templates']}")
        print(f"Query window: {summary['query_window_minutes']} minutes")
        print(f"Total daily logs: {summary['total_daily_logs']:,}")
        print(f"Total daily GB: {summary['total_daily_gb']:.2f}")

        print(f"\nTop {args.top} log messages by daily frequency:")
        for i, entry in enumerate(cx.get_top_offenders(args.top), 1):
            print(f"  {i:3d}. [{entry.daily_count:>12,}/day | {entry.daily_gb:.3f} GB] "
                  f"{entry.raw_message[:70]}")

        if args.output:
            with open(args.output, 'w') as f:
                json.dump(summary, f, indent=2)
            print(f"\nSummary saved to {args.output}")

    elif args.command == 'query':
        if args.type == 'frequency':
            query = generate_coralogix_query(args.app, args.subsystem)
            print(f"DataPrime frequency query for '{args.app}':")
            print(f"\n{query}\n")
            print("Run with: mcp_coralogix-server_get_logs")
            print(f"  start_date: <now - 15 minutes in ISO 8601>")
            print(f"  end_date: <now in ISO 8601>")
            print(f"  limit: 500")
        elif args.type == 'cost':
            query = generate_cost_query(args.app)
            print(f"DataPrime cost analysis query for '{args.app}':")
            print(f"\n{query}\n")
        elif args.type == 'units':
            query = generate_units_query(args.app)
            print(f"PromQL units query for '{args.app}':")
            print(f"\n{query}\n")
            print("Run with: mcp_coralogix-server_metrics__range_query")

    elif args.command == 'match':
        # Load Coralogix data
        cx = CoralogixData()
        if args.coralogix.endswith('.csv'):
            cx.load_from_csv(args.coralogix, query_minutes=args.query_minutes)
        else:
            cx.load_from_json(args.coralogix, query_minutes=args.query_minutes)

        # Load scan results
        with open(args.scan_results, 'r') as f:
            scan_data = json.load(f)

        logs = scan_data.get('logs', scan_data if isinstance(scan_data, list) else [])

        # Match
        results, summary = cx.match_to_code_logs(logs)

        print(f"\n{'='*60}")
        print(f"MATCHING RESULTS")
        print(f"{'='*60}")
        print(f"Match rate: {summary['match_rate']}%")
        print(f"Matched: {summary['matched']}/{summary['total_code_logs']}")
        print(f"By method: {summary['by_method']}")

        # Show top matched logs by frequency
        matched_results = [r for r in results if r.match_method != 'none']
        matched_results.sort(key=lambda r: r.daily_frequency, reverse=True)

        print(f"\nTop 15 matched logs by frequency:")
        for i, r in enumerate(matched_results[:15], 1):
            print(f"  {i:3d}. [{r.daily_frequency:>10,}/day] {r.code_file}:{r.code_line}")
            print(f"       Code: {r.code_message[:60]}")
            print(f"       Match: {r.match_method} ({r.match_score:.0%})")

        if args.output:
            output = {
                'summary': summary,
                'matches': [
                    {
                        'file': r.code_file,
                        'line': r.code_line,
                        'code_message': r.code_message,
                        'coralogix_message': r.coralogix_message,
                        'daily_frequency': r.daily_frequency,
                        'match_method': r.match_method,
                        'match_score': r.match_score
                    }
                    for r in results
                ]
            }
            with open(args.output, 'w') as f:
                json.dump(output, f, indent=2)
            print(f"\nResults saved to {args.output}")

    else:
        parser.print_help()
