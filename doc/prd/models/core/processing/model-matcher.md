# Matcher

## Overview
Evaluates a log line against filter conditions. A matcher can check for required substrings (contains) and excluded substrings (not_contains).

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| contains | Required substrings | list of text | No | ALL must be present (AND logic) |
| not_contains | Excluded substrings | list of text | No | NONE may be present (line rejected if any matches) |

## Evaluation Logic
1. If `contains` is set: the line must contain ALL listed strings
2. If `not_contains` is set: the line must NOT contain ANY listed string
3. Both conditions must pass for the line to match

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| Router | Belongs to | Part of router's filter chain |
| MatcherConfig | Configured by | Configuration source |
