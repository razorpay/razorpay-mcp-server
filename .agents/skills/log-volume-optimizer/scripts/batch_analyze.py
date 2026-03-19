#!/usr/bin/env python3
"""
Batch Log Volume Analyzer

Analyzes multiple Go repositories for log optimization opportunities.
Generates consolidated CSV report for Google Sheets sharing.

Usage:
    python batch_analyze.py --org razorpay --output-dir ./reports
    python batch_analyze.py --repos "pg-router,payments-upi" --output-dir ./reports
"""

import argparse
import csv
import json
import os
import subprocess
import tempfile
from dataclasses import dataclass, asdict
from datetime import datetime
from pathlib import Path
from typing import List, Dict, Optional
import shutil

# Import local modules
from scan_logs import scan_repository, scan_repository_all_languages, generate_summary, detect_language
from estimate_volume import VolumeEstimator


@dataclass
class RepoAnalysis:
    """Analysis result for a single repository."""
    repo_name: str
    application_name: str
    clone_path: str
    total_logs: int
    info_count: int
    error_count: int
    warn_count: int
    debug_count: int
    fatal_count: int
    logs_in_loops: int
    logs_in_error_handlers: int
    est_daily_units: float
    optimization_count: int
    est_savings_units: float
    est_savings_pct: float
    recommendations: List[Dict]
    status: str  # analyzed, skipped, error
    error_message: str = ""
    pr_url: str = ""

    def to_csv_row(self) -> Dict:
        return {
            'repo_name': self.repo_name,
            'application_name': self.application_name,
            'total_logs': self.total_logs,
            'info_count': self.info_count,
            'error_count': self.error_count,
            'warn_count': self.warn_count,
            'debug_count': self.debug_count,
            'fatal_count': self.fatal_count,
            'logs_in_loops': self.logs_in_loops,
            'logs_in_error_handlers': self.logs_in_error_handlers,
            'est_daily_units': round(self.est_daily_units, 2),
            'optimization_count': self.optimization_count,
            'est_savings_units': round(self.est_savings_units, 2),
            'est_savings_pct': round(self.est_savings_pct, 1),
            'status': self.status,
            'error_message': self.error_message,
            'pr_url': self.pr_url
        }


# Target applications from Coralogix query
TARGET_APPLICATIONS = [
    'ads-offers-engine', 'api', 'api-worker', 'asv', 'authz', 'barricade',
    'bin-service', 'charge-collections', 'checkout-affordability-api',
    'checkout-service', 'cmma', 'dashboard', 'downtime-manager', 'fts',
    'ledger', 'mozart', 'mozart-whitelisted', 'offers-engine', 'optimizer-core',
    'otpelfv2', 'payment-links', 'payment-methods', 'payments-card',
    'payments-nbplus', 'payments-upi', 'payouts', 'pg-router', 'reminders',
    'router', 'rto-prediction-service', 'scrooge', 'settlements', 'shield',
    'stork', 'terminals', 'tokens', 'ufh', 'ui-config-service', 'vault'
]


def run_command(cmd: List[str], cwd: str = None) -> tuple:
    """Run a shell command and return (success, output)."""
    try:
        result = subprocess.run(
            cmd,
            capture_output=True,
            text=True,
            cwd=cwd,
            timeout=300
        )
        return result.returncode == 0, result.stdout + result.stderr
    except subprocess.TimeoutExpired:
        return False, "Command timed out"
    except Exception as e:
        return False, str(e)


def get_go_repos_from_org(org: str, limit: int = 500) -> List[str]:
    """Fetch Go repositories from GitHub org."""
    cmd = [
        'gh', 'repo', 'list', org,
        '--limit', str(limit),
        '--json', 'name,primaryLanguage',
        '--jq', '.[] | select(.primaryLanguage.name == "Go") | .name'
    ]
    success, output = run_command(cmd)
    if success:
        repos = [r.strip() for r in output.strip().split('\n') if r.strip()]
        return repos
    return []


def filter_target_repos(repos: List[str]) -> List[str]:
    """Filter repos to only include target applications."""
    # Map repo names to application names (they often match)
    filtered = []
    for repo in repos:
        # Check if repo name matches any target application
        repo_lower = repo.lower().replace('-', '_').replace('_', '-')
        for app in TARGET_APPLICATIONS:
            app_normalized = app.lower().replace('_', '-')
            if repo_lower == app_normalized or repo.lower() == app.lower():
                filtered.append(repo)
                break
    return filtered


def clone_repo(org: str, repo: str, target_dir: str) -> tuple:
    """Clone a repository to target directory."""
    clone_url = f"git@github.com:{org}/{repo}.git"
    cmd = ['git', 'clone', '--depth', '1', clone_url, target_dir]
    return run_command(cmd)


def analyze_repo(
    repo_name: str,
    repo_path: str,
    avg_rps: float = 100,
    assigned_units: float = 500,
    language: str = 'auto'
) -> RepoAnalysis:
    """Analyze a single repository for log optimization."""

    # Scan for logs - auto-detect language or scan all
    if language == 'all':
        logs = scan_repository_all_languages(repo_path)
    else:
        logs = scan_repository(repo_path, language=language)

    if not logs:
        return RepoAnalysis(
            repo_name=repo_name,
            application_name=repo_name,
            clone_path=repo_path,
            total_logs=0,
            info_count=0,
            error_count=0,
            warn_count=0,
            debug_count=0,
            fatal_count=0,
            logs_in_loops=0,
            logs_in_error_handlers=0,
            est_daily_units=0,
            optimization_count=0,
            est_savings_units=0,
            est_savings_pct=0,
            recommendations=[],
            status='skipped',
            error_message='No log statements found'
        )

    # Generate summary
    summary = generate_summary(logs)

    # Estimate volume
    estimator = VolumeEstimator(avg_rps=avg_rps)
    log_dicts = [log.to_dict() for log in logs]
    estimates = estimator.estimate_all(log_dicts)
    report = estimator.generate_report(estimates, assigned_units)

    # Extract counts
    by_level = summary.get('by_level', {})

    # Generate recommendations
    recommendations = []
    potential_savings = 0

    for est in estimates:
        if est.optimization_potential in ['CONSOLIDATE_LOOP', 'CHANGE_TO_DEBUG', 'USE_METRICS', 'ADD_SAMPLING']:
            rec = {
                'file': est.file,
                'line': est.line,
                'level': est.level,
                'action': est.optimization_potential,
                'daily_units': est.daily_units,
                'message': est.message[:100]
            }
            recommendations.append(rec)

            # Calculate savings
            if est.optimization_potential == 'CONSOLIDATE_LOOP':
                potential_savings += est.daily_units * 0.9
            elif est.optimization_potential == 'CHANGE_TO_DEBUG':
                potential_savings += est.daily_units * 0.95
            elif est.optimization_potential == 'USE_METRICS':
                potential_savings += est.daily_units
            elif est.optimization_potential == 'ADD_SAMPLING':
                potential_savings += est.daily_units * 0.99

    total_units = report['summary']['total_daily_units']
    savings_pct = (potential_savings / total_units * 100) if total_units > 0 else 0

    return RepoAnalysis(
        repo_name=repo_name,
        application_name=repo_name,
        clone_path=repo_path,
        total_logs=len(logs),
        info_count=by_level.get('INFO', 0),
        error_count=by_level.get('ERROR', 0),
        warn_count=by_level.get('WARN', 0),
        debug_count=by_level.get('DEBUG', 0),
        fatal_count=by_level.get('FATAL', 0),
        logs_in_loops=summary.get('in_loops', 0),
        logs_in_error_handlers=summary.get('in_error_handlers', 0),
        est_daily_units=total_units,
        optimization_count=len(recommendations),
        est_savings_units=potential_savings,
        est_savings_pct=savings_pct,
        recommendations=recommendations,
        status='analyzed'
    )


def write_csv_report(analyses: List[RepoAnalysis], output_path: str):
    """Write consolidated CSV report."""
    if not analyses:
        return

    fieldnames = [
        'repo_name', 'application_name', 'total_logs', 'info_count',
        'error_count', 'warn_count', 'debug_count', 'fatal_count',
        'logs_in_loops', 'logs_in_error_handlers', 'est_daily_units',
        'optimization_count', 'est_savings_units', 'est_savings_pct',
        'status', 'error_message', 'pr_url'
    ]

    with open(output_path, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        for analysis in analyses:
            writer.writerow(analysis.to_csv_row())


def write_markdown_report(analyses: List[RepoAnalysis], output_path: str):
    """Write consolidated Markdown report."""
    with open(output_path, 'w') as f:
        f.write("# Log Volume Optimization - Consolidated Report\n\n")
        f.write(f"**Generated:** {datetime.now().isoformat()}\n\n")

        # Summary
        total_repos = len(analyses)
        analyzed = sum(1 for a in analyses if a.status == 'analyzed')
        total_savings = sum(a.est_savings_units for a in analyses)

        f.write("## Summary\n\n")
        f.write(f"- **Total Repositories Scanned:** {total_repos}\n")
        f.write(f"- **Repositories with Optimizations:** {analyzed}\n")
        f.write(f"- **Total Estimated Savings:** {total_savings:.2f} units/day\n\n")

        # Table
        f.write("## Repository Analysis\n\n")
        f.write("| Repo | Total Logs | INFO | ERROR | DEBUG | Daily Units | Optimizations | Savings | Status |\n")
        f.write("|------|------------|------|-------|-------|-------------|---------------|---------|--------|\n")

        for a in sorted(analyses, key=lambda x: x.est_savings_units, reverse=True):
            f.write(f"| {a.repo_name} | {a.total_logs} | {a.info_count} | {a.error_count} | {a.debug_count} | {a.est_daily_units:.1f} | {a.optimization_count} | {a.est_savings_units:.1f} ({a.est_savings_pct:.0f}%) | {a.status} |\n")

        f.write("\n")


def write_json_report(analyses: List[RepoAnalysis], output_path: str):
    """Write detailed JSON report with recommendations."""
    report = {
        'generated': datetime.now().isoformat(),
        'summary': {
            'total_repos': len(analyses),
            'analyzed': sum(1 for a in analyses if a.status == 'analyzed'),
            'total_logs': sum(a.total_logs for a in analyses),
            'total_daily_units': sum(a.est_daily_units for a in analyses),
            'total_savings_units': sum(a.est_savings_units for a in analyses),
        },
        'repositories': []
    }

    for a in analyses:
        repo_data = asdict(a) if hasattr(a, '__dataclass_fields__') else a.to_csv_row()
        repo_data['recommendations'] = a.recommendations
        report['repositories'].append(repo_data)

    with open(output_path, 'w') as f:
        json.dump(report, f, indent=2, default=str)


def main():
    parser = argparse.ArgumentParser(description='Batch analyze Go repositories for log optimization')
    parser.add_argument('--org', default='razorpay', help='GitHub organization')
    parser.add_argument('--repos', help='Comma-separated list of specific repos to analyze')
    parser.add_argument('--target-only', action='store_true',
                        help='Only analyze target applications from Coralogix query')
    parser.add_argument('--output-dir', default='./reports', help='Output directory for reports')
    parser.add_argument('--csv-output', default='consolidated-report.csv', help='CSV output filename')
    parser.add_argument('--avg-rps', type=float, default=100, help='Average RPS for estimation')
    parser.add_argument('--assigned-units', type=float, default=500, help='Assigned daily units quota')
    parser.add_argument('--keep-clones', action='store_true', help='Keep cloned repos after analysis')
    parser.add_argument('--clone-dir', help='Directory to clone repos (uses temp if not specified)')

    args = parser.parse_args()

    # Create output directory
    output_dir = Path(args.output_dir)
    output_dir.mkdir(parents=True, exist_ok=True)

    # Get repos to analyze
    if args.repos:
        repos = [r.strip() for r in args.repos.split(',')]
    else:
        print(f"Fetching Go repositories from {args.org}...")
        repos = get_go_repos_from_org(args.org)
        print(f"Found {len(repos)} Go repositories")

        if args.target_only:
            repos = filter_target_repos(repos)
            print(f"Filtered to {len(repos)} target applications")

    if not repos:
        print("No repositories to analyze")
        return

    # Setup clone directory
    if args.clone_dir:
        clone_base = Path(args.clone_dir)
        clone_base.mkdir(parents=True, exist_ok=True)
        temp_dir = None
    else:
        temp_dir = tempfile.mkdtemp(prefix='log-optimizer-')
        clone_base = Path(temp_dir)

    analyses = []

    try:
        for i, repo in enumerate(repos, 1):
            print(f"\n[{i}/{len(repos)}] Analyzing {repo}...")

            repo_path = clone_base / repo

            # Clone if needed
            if not repo_path.exists():
                print(f"  Cloning {repo}...")
                success, output = clone_repo(args.org, repo, str(repo_path))
                if not success:
                    print(f"  Failed to clone: {output[:100]}")
                    analyses.append(RepoAnalysis(
                        repo_name=repo,
                        application_name=repo,
                        clone_path=str(repo_path),
                        total_logs=0,
                        info_count=0,
                        error_count=0,
                        warn_count=0,
                        debug_count=0,
                        fatal_count=0,
                        logs_in_loops=0,
                        logs_in_error_handlers=0,
                        est_daily_units=0,
                        optimization_count=0,
                        est_savings_units=0,
                        est_savings_pct=0,
                        recommendations=[],
                        status='error',
                        error_message=f"Clone failed: {output[:100]}"
                    ))
                    continue

            # Analyze
            try:
                analysis = analyze_repo(
                    repo_name=repo,
                    repo_path=str(repo_path),
                    avg_rps=args.avg_rps,
                    assigned_units=args.assigned_units
                )
                analyses.append(analysis)
                print(f"  Found {analysis.total_logs} logs, {analysis.optimization_count} optimizations, ~{analysis.est_savings_units:.1f} units savings")
            except Exception as e:
                print(f"  Analysis failed: {e}")
                analyses.append(RepoAnalysis(
                    repo_name=repo,
                    application_name=repo,
                    clone_path=str(repo_path),
                    total_logs=0,
                    info_count=0,
                    error_count=0,
                    warn_count=0,
                    debug_count=0,
                    fatal_count=0,
                    logs_in_loops=0,
                    logs_in_error_handlers=0,
                    est_daily_units=0,
                    optimization_count=0,
                    est_savings_units=0,
                    est_savings_pct=0,
                    recommendations=[],
                    status='error',
                    error_message=str(e)[:200]
                ))

        # Write reports
        print("\n" + "="*50)
        print("Generating reports...")

        csv_path = output_dir / args.csv_output
        write_csv_report(analyses, str(csv_path))
        print(f"CSV report: {csv_path}")

        md_path = output_dir / args.csv_output.replace('.csv', '.md')
        write_markdown_report(analyses, str(md_path))
        print(f"Markdown report: {md_path}")

        json_path = output_dir / args.csv_output.replace('.csv', '.json')
        write_json_report(analyses, str(json_path))
        print(f"JSON report: {json_path}")

        # Print summary
        print("\n" + "="*50)
        print("SUMMARY")
        print("="*50)
        total_analyzed = sum(1 for a in analyses if a.status == 'analyzed')
        total_savings = sum(a.est_savings_units for a in analyses)
        print(f"Repositories analyzed: {total_analyzed}/{len(repos)}")
        print(f"Total potential savings: {total_savings:.2f} units/day")

        # Top repos by savings
        top_repos = sorted(
            [a for a in analyses if a.est_savings_units > 0],
            key=lambda x: x.est_savings_units,
            reverse=True
        )[:10]

        if top_repos:
            print("\nTop 10 repos by potential savings:")
            for r in top_repos:
                print(f"  {r.repo_name}: {r.est_savings_units:.1f} units ({r.est_savings_pct:.0f}%)")

    finally:
        # Cleanup
        if temp_dir and not args.keep_clones:
            print(f"\nCleaning up temp directory: {temp_dir}")
            shutil.rmtree(temp_dir, ignore_errors=True)
        elif args.keep_clones:
            print(f"\nCloned repos kept at: {clone_base}")


if __name__ == '__main__':
    main()
