# resume-gen

Генерує PDF-резюме з `resume.yaml` у тому самому LaTeX-стилі.

## Вимоги

- Go 1.22+
- `pdflatex` у PATH
  - macOS: `brew install --cask mactex-no-gui` (або BasicTeX)
  - Ubuntu/Debian: `sudo apt install texlive-latex-base texlive-latex-recommended`

## Перший запуск

```sh
go mod tidy        # один раз — підтягне gopkg.in/yaml.v3 і створить go.sum
go run .           # resume.yaml -> resume.pdf
```

## Флаги

```sh
go run . -in resume.yaml -out my_name_resume.pdf
go run . -tex      # додатково зберегти згенерований .tex поруч із PDF
```

## Як додати проєкт

Додай елемент у `projects:` у `resume.yaml` (порядок у файлі = порядок у PDF):

```yaml
  - name: MyProject(Role)
    dates: Jan 2027 - Present
    stack: [Go, PostgreSQL]
    url: https://example.com          # необов'язково, вирівнюється праворуч
    description: >                    # необов'язково
      Короткий опис проєкту.
    responsibilities:                 # необов'язково; якщо порожньо — без "Responsibilities:"
      - "Перший пункт."
      - "Другий пункт."
```

Пиши звичайний текст: символи `& % $ # _ ~ — ²` екрануються автоматично.
Пункти з `: ` всередині бери в лапки. Помилка в назві поля (наприклад `respnsibilities`) зупинить генерацію з повідомленням.

Верстка (відступи, шрифти) — у `template.tex`, він вбудовується в бінарник через `go:embed`.
