#!/usr/bin/env python3
"""
Log Volume Estimator for Log Volume Optimizer Skill

Estimates daily log volume and Coralogix units based on:
- Scanned log statements
- Traffic parameters (RPS, merchants)
- Log characteristics (level, context, size)

Usage:
    python estimate_volume.py logs.json --rps 500 --output estimates.json
"""

import argparse
import json
from dataclasses import asdict, dataclass
from typing import Dict, List, Optional


@dataclass
class LogVolumeEstimate:
    """Volume estimate for a single log statement."""
    file: str
    line: int
    level: str
    function: str
    message: str
    
    # Volume factors
    trigger_probability: float
    rps: float
    log_size_bytes: int
    
    # Calculated values
    daily_invocations: float
    daily_bytes: float
    daily_units: float
    
    # Impact category
    impact_category: str  # CRITICAL, HIGH, MEDIUM, LOW
    
    # Flags
    in_loop: bool
    in_error_handler: bool
    optimization_potential: str
    
    def to_dict(self) -> Dict:
        return asdict(self)


class VolumeEstimator:
    """Estimates log volume based on traffic parameters."""
    
    # Coralogix pricing: 1 unit = 1 MB
    BYTES_PER_UNIT = 1_000_000
    
    # Seconds per day
    SECONDS_PER_DAY = 86400
    
    # Default trigger probabilities
    DEFAULT_TRIGGER_PROB = 1.0
    ERROR_HANDLER_PROB = 0.01  # 1% error rate
    CONDITIONAL_PROB = 0.5     # 50% branch probability
    
    # Impact thresholds (units/day)
    CRITICAL_THRESHOLD = 100
    HIGH_THRESHOLD = 10
    MEDIUM_THRESHOLD = 1
    
    def __init__(
        self,
        avg_rps: float = 100,
        peak_rps: float = 500,
        error_rate: float = 0.01,
        loop_multiplier: float = 10
    ):
        """
        Initialize estimator with traffic parameters.
        
        Args:
            avg_rps: Average requests per second
            peak_rps: Peak requests per second
            error_rate: Assumed error rate (0.0 to 1.0)
            loop_multiplier: Average loop iterations
        """
        self.avg_rps = avg_rps
        self.peak_rps = peak_rps
        self.error_rate = error_rate
        self.loop_multiplier = loop_multiplier
    
    def estimate_log(
        self,
        log: Dict,
        route_rps: Optional[float] = None
    ) -> LogVolumeEstimate:
        """
        Estimate volume for a single log statement.
        
        Args:
            log: Log statement dict from scanner
            route_rps: Optional route-specific RPS (defaults to avg_rps)
        
        Returns:
            LogVolumeEstimate with calculated values
        """
        rps = route_rps or self.avg_rps
        
        # Calculate trigger probability
        trigger_prob = self._calculate_trigger_probability(log)
        
        # Adjust for loops
        effective_rps = rps
        if log.get('in_loop', False):
            effective_rps = rps * self.loop_multiplier
        
        # Calculate daily invocations
        daily_invocations = effective_rps * self.SECONDS_PER_DAY * trigger_prob
        
        # Get log size
        log_size = log.get('estimated_size_bytes', 300)
        
        # Calculate daily volume
        daily_bytes = daily_invocations * log_size
        daily_units = daily_bytes / self.BYTES_PER_UNIT
        
        # Determine impact category
        impact_category = self._categorize_impact(daily_units)
        
        # Determine optimization potential
        optimization_potential = self._assess_optimization(log, daily_units)
        
        return LogVolumeEstimate(
            file=log.get('file', ''),
            line=log.get('line', 0),
            level=log.get('level', 'INFO'),
            function=log.get('function', 'unknown'),
            message=log.get('message', ''),
            trigger_probability=trigger_prob,
            rps=effective_rps,
            log_size_bytes=log_size,
            daily_invocations=round(daily_invocations, 2),
            daily_bytes=round(daily_bytes, 2),
            daily_units=round(daily_units, 4),
            impact_category=impact_category,
            in_loop=log.get('in_loop', False),
            in_error_handler=log.get('in_error_handler', False),
            optimization_potential=optimization_potential
        )
    
    def _calculate_trigger_probability(self, log: Dict) -> float:
        """Calculate probability that log is triggered per request."""
        level = log.get('level', 'INFO').upper()
        in_error_handler = log.get('in_error_handler', False)
        
        # Error handler logs only trigger on errors
        if in_error_handler:
            return self.error_rate
        
        # ERROR/FATAL logs typically only on errors
        if level in ('ERROR', 'FATAL', 'PANIC'):
            return self.error_rate
        
        # WARN logs somewhat less frequent
        if level == 'WARN':
            return 0.1  # 10% of requests
        
        # DEBUG logs typically disabled in prod
        if level == 'DEBUG':
            return 0.0  # Assume disabled
        
        # INFO logs always trigger (in their code path)
        return self.DEFAULT_TRIGGER_PROB
    
    def _categorize_impact(self, daily_units: float) -> str:
        """Categorize log by daily unit impact."""
        if daily_units >= self.CRITICAL_THRESHOLD:
            return 'CRITICAL'
        elif daily_units >= self.HIGH_THRESHOLD:
            return 'HIGH'
        elif daily_units >= self.MEDIUM_THRESHOLD:
            return 'MEDIUM'
        else:
            return 'LOW'
    
    def _assess_optimization(self, log: Dict, daily_units: float) -> str:
        """Assess optimization potential for this log."""
        level = log.get('level', 'INFO').upper()
        in_loop = log.get('in_loop', False)
        in_error_handler = log.get('in_error_handler', False)
        message = log.get('message', '').lower()
        
        # Logs in loops are prime optimization targets
        if in_loop:
            return 'CONSOLIDATE_LOOP'
        
        # Entry/exit logs should be DEBUG
        if any(word in message for word in ['entering', 'entered', 'exiting', 'exit', 'starting']):
            if level == 'INFO':
                return 'CHANGE_TO_DEBUG'
        
        # Success logs could be metrics
        if any(word in message for word in ['success', 'completed', 'done']):
            if daily_units > 10:
                return 'USE_METRICS'
        
        # High volume INFO logs could be sampled
        if level == 'INFO' and daily_units > 50:
            return 'ADD_SAMPLING'
        
        # Error logs in error handlers are usually fine
        if level in ('ERROR', 'FATAL') and in_error_handler:
            return 'KEEP'
        
        # Low impact logs are fine
        if daily_units < 1:
            return 'KEEP'
        
        return 'REVIEW'
    
    def estimate_all(
        self,
        logs: List[Dict],
        route_rps_map: Optional[Dict[str, float]] = None
    ) -> List[LogVolumeEstimate]:
        """
        Estimate volume for all log statements.
        
        Args:
            logs: List of log dicts from scanner
            route_rps_map: Optional mapping of function/route to RPS
        
        Returns:
            List of estimates sorted by daily_units descending
        """
        estimates = []
        
        for log in logs:
            # Get route-specific RPS if available
            route_rps = None
            if route_rps_map:
                function = log.get('function', '')
                route_rps = route_rps_map.get(function)
            
            estimate = self.estimate_log(log, route_rps)
            estimates.append(estimate)
        
        # Sort by daily_units descending
        estimates.sort(key=lambda e: e.daily_units, reverse=True)
        
        return estimates
    
    def generate_report(
        self,
        estimates: List[LogVolumeEstimate],
        assigned_units: float = 0
    ) -> Dict:
        """
        Generate a comprehensive volume report.
        
        Args:
            estimates: List of volume estimates
            assigned_units: Daily quota (0 if unknown)
        
        Returns:
            Report dict with summary and details
        """
        total_units = sum(e.daily_units for e in estimates)
        total_bytes = sum(e.daily_bytes for e in estimates)
        
        # Count by category
        by_category = {}
        for cat in ['CRITICAL', 'HIGH', 'MEDIUM', 'LOW']:
            cat_estimates = [e for e in estimates if e.impact_category == cat]
            by_category[cat] = {
                'count': len(cat_estimates),
                'units': round(sum(e.daily_units for e in cat_estimates), 2),
                'percent': 0
            }
            if total_units > 0:
                by_category[cat]['percent'] = round(
                    by_category[cat]['units'] / total_units * 100, 1
                )
        
        # Count by level
        by_level = {}
        for level in ['DEBUG', 'INFO', 'WARN', 'ERROR', 'FATAL']:
            level_estimates = [e for e in estimates if e.level == level]
            by_level[level] = {
                'count': len(level_estimates),
                'units': round(sum(e.daily_units for e in level_estimates), 2)
            }
        
        # Count optimization opportunities
        optimization_counts = {}
        for e in estimates:
            pot = e.optimization_potential
            optimization_counts[pot] = optimization_counts.get(pot, 0) + 1
        
        # Calculate potential savings
        potential_savings = 0
        for e in estimates:
            if e.optimization_potential == 'CONSOLIDATE_LOOP':
                potential_savings += e.daily_units * 0.9  # 90% reduction
            elif e.optimization_potential == 'CHANGE_TO_DEBUG':
                potential_savings += e.daily_units * 0.95  # 95% reduction (disabled)
            elif e.optimization_potential == 'USE_METRICS':
                potential_savings += e.daily_units * 1.0  # 100% reduction
            elif e.optimization_potential == 'ADD_SAMPLING':
                potential_savings += e.daily_units * 0.99  # 99% reduction
        
        report = {
            'summary': {
                'total_logs': len(estimates),
                'total_daily_units': round(total_units, 2),
                'total_daily_bytes': round(total_bytes, 2),
                'total_daily_gb': round(total_bytes / 1_000_000_000, 3),
            },
            'quota': {
                'assigned_units': assigned_units,
                'utilization_percent': round(total_units / assigned_units * 100, 1) if assigned_units > 0 else 0,
                'remaining_units': round(assigned_units - total_units, 2) if assigned_units > 0 else 0,
            },
            'by_category': by_category,
            'by_level': by_level,
            'optimization': {
                'opportunities': optimization_counts,
                'potential_savings_units': round(potential_savings, 2),
                'potential_savings_percent': round(potential_savings / total_units * 100, 1) if total_units > 0 else 0,
            },
            'top_logs': [e.to_dict() for e in estimates[:20]],  # Top 20 by volume
        }
        
        return report


def main():
    parser = argparse.ArgumentParser(description='Estimate log volume')
    parser.add_argument('logs_file', help='Path to logs.json from scanner')
    parser.add_argument('--rps', type=float, default=100, help='Average RPS')
    parser.add_argument('--peak-rps', type=float, default=500, help='Peak RPS')
    parser.add_argument('--assigned-units', type=float, default=0, help='Daily quota')
    parser.add_argument('--error-rate', type=float, default=0.01, help='Error rate')
    parser.add_argument('--loop-multiplier', type=float, default=10, help='Avg loop iterations')
    parser.add_argument('--output', '-o', default='estimates.json', help='Output file')
    
    args = parser.parse_args()
    
    # Load logs
    with open(args.logs_file, 'r') as f:
        data = json.load(f)
    
    logs = data.get('logs', [])
    print(f"Loaded {len(logs)} log statements")
    
    # Create estimator
    estimator = VolumeEstimator(
        avg_rps=args.rps,
        peak_rps=args.peak_rps,
        error_rate=args.error_rate,
        loop_multiplier=args.loop_multiplier
    )
    
    # Estimate all logs
    estimates = estimator.estimate_all(logs)
    
    # Generate report
    report = estimator.generate_report(estimates, args.assigned_units)
    
    # Print summary
    print("\n=== Volume Estimate Summary ===")
    print(f"Total logs: {report['summary']['total_logs']}")
    print(f"Estimated daily units: {report['summary']['total_daily_units']}")
    print(f"Estimated daily volume: {report['summary']['total_daily_gb']:.3f} GB")
    
    if args.assigned_units > 0:
        print(f"\nQuota utilization: {report['quota']['utilization_percent']}%")
        print(f"Remaining units: {report['quota']['remaining_units']}")
    
    print(f"\nBy category:")
    for cat, data in report['by_category'].items():
        print(f"  {cat}: {data['count']} logs, {data['units']} units ({data['percent']}%)")
    
    print(f"\nOptimization potential:")
    print(f"  Potential savings: {report['optimization']['potential_savings_units']} units")
    print(f"  Reduction: {report['optimization']['potential_savings_percent']}%")
    
    # Write output
    output = {
        'report': report,
        'estimates': [e.to_dict() for e in estimates]
    }
    
    with open(args.output, 'w') as f:
        json.dump(output, f, indent=2)
    
    print(f"\nFull report written to {args.output}")


if __name__ == '__main__':
    main()
