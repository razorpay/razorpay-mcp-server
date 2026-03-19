#!/usr/bin/env python3
"""
Decision Engine for Log Volume Optimizer

Implements the consolidated decision rules from team approaches (Monark, Ankur, Rajeev).
Each log statement is evaluated against rules in priority order - first match wins.

Usage:
    from decision_engine import DecisionEngine

    engine = DecisionEngine()
    decision = engine.decide(log_entry, frequency=1000000)
"""

import re
from dataclasses import dataclass
from enum import Enum
from typing import Dict, List, Optional, Tuple


class Action(Enum):
    """Actions that can be taken on a log statement."""
    KEEP = "KEEP"
    DOWNGRADE_TO_DEBUG = "DOWNGRADE_TO_DEBUG"
    REMOVE = "REMOVE"
    AGGREGATE = "AGGREGATE"
    SAMPLE = "SAMPLE"
    USE_METRIC = "USE_METRIC"
    FLAG_REVIEW = "FLAG_REVIEW"


@dataclass
class Decision:
    """Result of evaluating a log statement."""
    action: Action
    rule_id: int
    reason: str
    confidence: str  # high, medium, low
    estimated_savings_pct: float  # 0-100


class DecisionEngine:
    """
    Evaluates log statements and decides what action to take.

    Rules are evaluated in priority order (lower number = higher priority).
    First matching rule wins.
    """

    # Patterns for rule matching
    ERROR_KEYWORDS = ['failed', 'error', 'timeout', 'panic', 'exception', 'fatal', 'crash']
    BUSINESS_KEYWORDS = ['payment', 'transaction', 'settlement', 'refund', 'order', 'invoice', 'payout']
    ENTRY_EXIT_PATTERNS = [
        r'\bentering\b', r'\bexiting\b', r'\bstarting\b', r'\bcompleted\b',
        r'\bbegin\b', r'\bend\b', r'\bstart\b', r'\bfinish\b', r'\binit\b',
        r'\binitialized\b', r'\bshutdown\b', r'\bstarted\b', r'\bstopped\b'
    ]
    SUCCESS_PATTERNS = [
        r'\bsuccessfully\b', r'\bsuccess\b', r'\bdone\b', r'\bcomplete\b',
        r'\bfinished\b', r'\bprocessed\b', r'\bhandled\b'
    ]
    METRICS_PATTERNS = [
        r'\bcount\b', r'\btotal\b', r'\bprocessed\s+\d+', r'\bduration\b',
        r'\blatency\b', r'\btook\s+\d+', r'\bms\b', r'\btime\b', r'\bthroughput\b',
        r'\brequests?\s+per\b', r'\brate\b'
    ]
    SECRETS_PATTERNS = [
        r'\bpassword\b', r'\btoken\b', r'\bsecret\b', r'\bkey\b', r'\bcredential\b',
        r'\bapi_key\b', r'\baccess_token\b', r'\bauth\b', r'\bprivate\b'
    ]

    # Frequency thresholds (per day)
    HIGH_FREQ_THRESHOLD = 1_000_000  # 1M/day
    LOOP_AGGREGATE_THRESHOLD = 100_000  # 100K/day
    LOOP_SAMPLE_THRESHOLD = 10_000  # 10K/day

    def __init__(self):
        self.rules = self._build_rules()

    def _build_rules(self) -> List[Tuple[int, str, callable]]:
        """Build the ordered list of decision rules."""
        return [
            (1, "ERROR/FATAL level always kept", self._rule_error_fatal),
            (2, "WARN level always kept", self._rule_warn),
            (3, "In error handler context", self._rule_error_handler),
            (4, "Message contains error keywords", self._rule_error_keywords),
            (5, "Business-critical event", self._rule_business_critical),
            (6, "In loop with very high frequency - aggregate", self._rule_loop_aggregate),
            (7, "In loop with high frequency - sample", self._rule_loop_sample),
            (8, "Entry/exit pattern", self._rule_entry_exit),
            (9, "Success confirmation pattern", self._rule_success_pattern),
            (10, "High-frequency INFO log", self._rule_high_freq_info),
            (11, "Metrics pattern - convert to metric", self._rule_metrics_pattern),
            (12, "Contains secrets/PII", self._rule_secrets),
            (13, "DEBUG level already", self._rule_debug),
            (14, "Default - keep", self._rule_default),
        ]

    def decide(self, log_entry: Dict, frequency: int = 0) -> Decision:
        """
        Evaluate a log entry and return a decision.

        Args:
            log_entry: Dict with keys: level, message, in_loop, in_error_handler, file, line
            frequency: Daily frequency from Coralogix (0 if unknown)

        Returns:
            Decision object with action, reason, confidence
        """
        for rule_id, rule_name, rule_func in self.rules:
            result = rule_func(log_entry, frequency)
            if result:
                action, reason, confidence, savings = result
                return Decision(
                    action=action,
                    rule_id=rule_id,
                    reason=reason,
                    confidence=confidence,
                    estimated_savings_pct=savings
                )

        # Should never reach here due to default rule
        return Decision(
            action=Action.KEEP,
            rule_id=14,
            reason="Default: no matching rule",
            confidence="low",
            estimated_savings_pct=0
        )

    def _get_level(self, log_entry: Dict) -> str:
        """Extract normalized log level."""
        return log_entry.get('level', '').upper()

    def _get_message(self, log_entry: Dict) -> str:
        """Extract log message."""
        return log_entry.get('message', '').lower()

    def _matches_patterns(self, text: str, patterns: List[str]) -> bool:
        """Check if text matches any of the patterns."""
        text_lower = text.lower()
        for pattern in patterns:
            if re.search(pattern, text_lower, re.IGNORECASE):
                return True
        return False

    def _contains_keywords(self, text: str, keywords: List[str]) -> bool:
        """Check if text contains any of the keywords."""
        text_lower = text.lower()
        return any(kw in text_lower for kw in keywords)

    # Rule implementations

    def _rule_error_fatal(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 1: ERROR and FATAL logs are always kept."""
        level = self._get_level(log_entry)
        if level in ['ERROR', 'FATAL']:
            return (Action.KEEP, f"{level} level must be preserved", "high", 0)
        return None

    def _rule_warn(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 2: WARN logs are always kept."""
        if self._get_level(log_entry) == 'WARN':
            return (Action.KEEP, "WARN level indicates potential issues", "high", 0)
        return None

    def _rule_error_handler(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 3: Logs in error handler context are kept."""
        if log_entry.get('in_error_handler', False):
            return (Action.KEEP, "In error handler - context needed for debugging", "high", 0)
        return None

    def _rule_error_keywords(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 4: Messages with error semantics are kept."""
        message = self._get_message(log_entry)
        if self._contains_keywords(message, self.ERROR_KEYWORDS):
            return (Action.KEEP, "Message indicates error condition", "high", 0)
        return None

    def _rule_business_critical(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 5: Business-critical events are kept."""
        message = self._get_message(log_entry)
        if self._contains_keywords(message, self.BUSINESS_KEYWORDS):
            return (Action.KEEP, "Business-critical event logging", "high", 0)
        return None

    def _rule_loop_aggregate(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 6: Very high frequency logs in loops should be aggregated."""
        if log_entry.get('in_loop', False) and frequency >= self.LOOP_AGGREGATE_THRESHOLD:
            savings = 90  # Aggregating typically saves 90%+
            return (
                Action.AGGREGATE,
                f"In loop with {frequency:,}/day - aggregate to summary",
                "high",
                savings
            )
        return None

    def _rule_loop_sample(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 7: High frequency logs in loops should be sampled."""
        if log_entry.get('in_loop', False) and frequency >= self.LOOP_SAMPLE_THRESHOLD:
            savings = 99  # 1% sampling saves 99%
            return (
                Action.SAMPLE,
                f"In loop with {frequency:,}/day - add 1% sampling",
                "medium",
                savings
            )
        return None

    def _rule_entry_exit(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 8: Entry/exit logs should be DEBUG."""
        message = self._get_message(log_entry)
        level = self._get_level(log_entry)
        if level == 'INFO' and self._matches_patterns(message, self.ENTRY_EXIT_PATTERNS):
            savings = 95  # DEBUG typically not shipped to Coralogix
            return (
                Action.DOWNGRADE_TO_DEBUG,
                "Entry/exit log - not needed in production",
                "high",
                savings
            )
        return None

    def _rule_success_pattern(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 9: Success confirmations should be DEBUG."""
        message = self._get_message(log_entry)
        level = self._get_level(log_entry)
        if level == 'INFO' and self._matches_patterns(message, self.SUCCESS_PATTERNS):
            # Don't downgrade if it also contains business keywords
            if not self._contains_keywords(message, self.BUSINESS_KEYWORDS):
                savings = 95
                return (
                    Action.DOWNGRADE_TO_DEBUG,
                    "Success confirmation - absence indicates failure",
                    "high",
                    savings
                )
        return None

    def _rule_high_freq_info(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 10: High-frequency INFO logs should be DEBUG."""
        level = self._get_level(log_entry)
        if level == 'INFO' and frequency >= self.HIGH_FREQ_THRESHOLD:
            savings = 95
            return (
                Action.DOWNGRADE_TO_DEBUG,
                f"High frequency INFO ({frequency:,}/day) - too verbose",
                "medium",
                savings
            )
        return None

    def _rule_metrics_pattern(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 11: Metrics-like logs should be Prometheus metrics."""
        message = self._get_message(log_entry)
        if self._matches_patterns(message, self.METRICS_PATTERNS):
            savings = 100  # Metrics don't generate logs
            return (
                Action.USE_METRIC,
                "Counting/timing pattern - should be Prometheus metric",
                "medium",
                savings
            )
        return None

    def _rule_secrets(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 12: Logs with potential secrets need review."""
        message = self._get_message(log_entry)
        if self._matches_patterns(message, self.SECRETS_PATTERNS):
            return (
                Action.FLAG_REVIEW,
                "Potential secrets/PII in log - needs human review",
                "high",
                0
            )
        return None

    def _rule_debug(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 13: DEBUG logs are already at lowest level."""
        if self._get_level(log_entry) == 'DEBUG':
            return (Action.KEEP, "Already DEBUG level", "high", 0)
        return None

    def _rule_default(self, log_entry: Dict, frequency: int) -> Optional[Tuple]:
        """Rule 14: Default to keeping the log."""
        return (Action.KEEP, "No optimization rule matched - keeping", "low", 0)


def batch_decide(logs: List[Dict], frequencies: Dict[str, int] = None,
                 coralogix_data=None) -> List[Dict]:
    """
    Process a batch of logs and return decisions.

    Args:
        logs: List of log entries
        frequencies: Dict mapping message templates to daily frequency
        coralogix_data: Optional CoralogixData instance for real production frequencies
                       (takes priority over frequencies dict)

    Returns:
        List of log entries with 'decision' field added
    """
    engine = DecisionEngine()
    frequencies = frequencies or {}

    results = []
    freq_sources = {'coralogix': 0, 'manual': 0, 'none': 0}

    for log in logs:
        freq = 0
        source = 'none'

        # Priority 1: Real Coralogix data
        if coralogix_data and hasattr(coralogix_data, 'get_daily_frequency'):
            message = log.get('message', '') or log.get('message_template', '')
            real_freq = coralogix_data.get_daily_frequency(message)
            if real_freq > 0:
                freq = real_freq
                source = 'coralogix'

        # Priority 2: Manual frequency dict
        if freq == 0 and frequencies:
            message = log.get('message', '')[:100]
            freq = frequencies.get(message, 0)
            if freq > 0:
                source = 'manual'

        freq_sources[source] = freq_sources.get(source, 0) + 1
        decision = engine.decide(log, frequency=freq)

        log_with_decision = log.copy()
        log_with_decision['decision'] = {
            'action': decision.action.value,
            'rule_id': decision.rule_id,
            'reason': decision.reason,
            'confidence': decision.confidence,
            'estimated_savings_pct': decision.estimated_savings_pct,
            'frequency': freq,
            'frequency_source': source
        }
        results.append(log_with_decision)

    return results


def summarize_decisions(logs_with_decisions: List[Dict]) -> Dict:
    """
    Summarize decisions for reporting.

    Returns:
        Dict with action counts and estimated savings
    """
    summary = {
        'total': len(logs_with_decisions),
        'by_action': {},
        'by_confidence': {'high': 0, 'medium': 0, 'low': 0},
        'actionable': 0,
        'estimated_total_savings_pct': 0
    }

    total_savings = 0
    actionable_count = 0

    for log in logs_with_decisions:
        decision = log.get('decision', {})
        action = decision.get('action', 'KEEP')
        confidence = decision.get('confidence', 'low')
        savings = decision.get('estimated_savings_pct', 0)

        # Count by action
        summary['by_action'][action] = summary['by_action'].get(action, 0) + 1

        # Count by confidence
        summary['by_confidence'][confidence] = summary['by_confidence'].get(confidence, 0) + 1

        # Count actionable (not KEEP)
        if action != 'KEEP':
            actionable_count += 1
            total_savings += savings

    summary['actionable'] = actionable_count
    if actionable_count > 0:
        summary['estimated_total_savings_pct'] = round(total_savings / actionable_count, 1)

    return summary


# CLI interface for testing
if __name__ == '__main__':
    import argparse
    import json

    parser = argparse.ArgumentParser(description='Test decision engine with log entries')
    parser.add_argument('--input', required=True, help='JSON file with log entries')
    parser.add_argument('--frequencies', help='JSON file with message:frequency mapping')
    parser.add_argument('--output', help='Output file for decisions')

    args = parser.parse_args()

    # Load logs
    with open(args.input, 'r') as f:
        logs = json.load(f)

    # Load frequencies if provided
    frequencies = {}
    if args.frequencies:
        with open(args.frequencies, 'r') as f:
            frequencies = json.load(f)

    # Process
    results = batch_decide(logs, frequencies)
    summary = summarize_decisions(results)

    # Output
    output = {
        'summary': summary,
        'logs': results
    }

    if args.output:
        with open(args.output, 'w') as f:
            json.dump(output, f, indent=2)
        print(f"Results written to {args.output}")
    else:
        print(json.dumps(output, indent=2))

    # Print summary
    print(f"\n{'='*50}")
    print("DECISION SUMMARY")
    print(f"{'='*50}")
    print(f"Total logs: {summary['total']}")
    print(f"Actionable: {summary['actionable']}")
    print(f"\nBy Action:")
    for action, count in sorted(summary['by_action'].items()):
        print(f"  {action}: {count}")
    print(f"\nEstimated savings: {summary['estimated_total_savings_pct']}%")
