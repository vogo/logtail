# MatcherConfig

## Overview
Configuration for a log line filter condition. Multiple matchers on a router use AND logic.

## Attributes

| Attribute | Description | Type | Required | Notes |
|-----------|-------------|------|----------|-------|
| contains | Strings that must ALL be present in the line | list of text | No | AND logic within the list |
| not_contains | Strings that must NOT be present in the line | list of text | No | Line is rejected if ANY string matches |

## Relationships

| Related Model | Relationship Type | Description |
|---------------|-------------------|-------------|
| RouterConfig | Belongs to | Part of router configuration |
