# cross-locale-display-search Specification

## Purpose
TBD - created by archiving change promptrepo-cross-locale-search-metadata-v1. Update Purpose after archive.
## Requirements
### Requirement: Display metadata from every locale is searchable
Search SHALL consider title, summary and aliases from every solution locale while returning the result card and exact ref in the selected locale.

#### Scenario: Chinese alias finds an English Agent template
- **WHEN** a solution has an English compilable template and Chinese review metadata and the request uses `locale=en` with a Chinese alias
- **THEN** search returns the English result card and exact ref with the localized alias match score

### Requirement: Existing selected-locale scoring remains stable
Cross-locale matching SHALL not lower or double-count the score produced by the selected locale.

#### Scenario: Existing Chinese compatibility fixture
- **WHEN** the private v0.1.0 search fixture is queried in `zh-CN`
- **THEN** its result, compatibility and score remain unchanged

