# 🌍 Localization System

This system handles multi-language support for the bot using YAML files and Go's `embed` package.

## 📂 Project Structure

* `internal/locales/*.yml`: Language translation files.
* `internal/locales/loader.go`: Core logic for loading and fetching strings.
* `internal/modules/helpers.go`: Contains `F()` and `FWithLang()` for easy access.

## 🛠️ How to Add a New Language

1. **Create File:** Create `internal/locales/{code}.yml` (e.g., `es.yml`).
2. **Define Name:** Set the display name: `name: "🇪🇸 Español"`.
3. **Translate:** Copy keys from `en.yml` and translate.
    * Use `{key}` for variables (e.g., `Hello {user}`).
    * Use `|` for multi-line strings.
4. **Deploy:** Files are automatically embedded; no code changes are required.

## 📝 Usage for Developers

Use these helpers instead of calling the loader directly:

* `F(chatID, key, args...)`: Automatically detects chat language.
* `FWithLang(lang, key, args...)`: Force a specific language.

---

## 👥 Language Contributors

| Language     | Code | Contributor                                              |
|:-------------|:-----|:---------------------------------------------------------|
| 🇺🇸 English | `en` | [@gabrielmaialva33](https://github.com/gabrielmaialva33) |
| 🇮🇳 Hindi   | `hi` | [@gabrielmaialva33](https://github.com/gabrielmaialva33) |
| 🇸🇦 Arabic  | `ar` | [@gabrielmaialva33](https://github.com/gabrielmaialva33) |
| 🇹🇷 Turkish | `tr` | [@gabrielmaialva33](https://github.com/gabrielmaialva33) |

---
*All YAML files in `internal/locales/` are automatically embedded into the binary during compilation.*