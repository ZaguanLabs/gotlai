# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-12-18

### Added

- **Translation Styles**: Control tone and formality with 5 styles:
  - `StyleFormal` - Professional language for official documents
  - `StyleNeutral` - Balanced tone for general content (default)
  - `StyleCasual` - Conversational language for blogs/social media
  - `StyleMarketing` - Persuasive language for promotional content
  - `StyleTechnical` - Precise language for technical documentation

- **Glossary Support**: Provide preferred translations for specific terms
  - `WithGlossary(map[string]string)` option for consistent terminology
  - Glossary entries are included in AI prompts

- **Enhanced Language Support**:
  - Expanded short code mappings (70+ codes including country codes like `jp`, `br`, `tw`)
  - Locale clarifications for language variants (Norwegian Bokmål/Nynorsk, Chinese Simplified/Traditional, etc.)
  - Better handling of regional variants (pt_BR vs pt_PT, es_ES vs es_MX)

- **New Translator Methods**:
  - `IsSourceLang()` - Check if target matches source language
  - `IsRTL()` - Check if target language is right-to-left
  - `GetDir()` - Get text direction ("ltr" or "rtl")
  - `Glossary()` - Get configured glossary
  - `Style()` - Get configured style
  - `Context()` - Get configured context
  - `ExcludedTerms()` - Get excluded terms list

- **Improved AI Prompts**:
  - Enhanced system prompt with style descriptions
  - Locale-specific hints for better translations
  - Quality check instructions for native-sounding output
  - Context hint handling (`{{__ctx__:...}}` format)

### Changed

- Default translation style is now `StyleNeutral`
- OpenAI provider uses improved prompt engineering for higher quality translations
- `TranslateRequest` now includes `Glossary` and `Style` fields

## [0.1.0] - Initial Release

### Added

- Core translation engine with OpenAI provider
- HTML and Go source code processors
- In-memory and Redis caching
- RTL language support
- Retry logic with exponential backoff
- Rate limiting
- CLI tool for command-line translation
- Diff mode for incremental updates
