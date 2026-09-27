# Regression fixture
## Parameters

| Параметр | Тип | По умолчанию | Описание |
|---|---|---|---|
| `--width` | int | 80 | Максимальная ширина вывода в колонках терминала, после которой текст переносится |
| `--theme` | string | auto | Цветовая тема: dark, light, auto, или путь к JSON-файлу с темой |

## Inline content, emoji, CJK, alignment

| Статус | Имя | Ссылка | Число |
|:---:|---|---|---:|
| ✅ | **жирный** | [docs](https://example.com/very/long/path/to/documentation/page) | 1 |
| ⚠️ | 日本語テキスト | `code \| pipe` | 1000 |
| 👨‍👩‍👧 | обычный | - | 42 |

## Long sequence diagram

```mermaid
sequenceDiagram
    autonumber
    participant B as Browser
    participant G as Gateway
    participant A as Auth
    participant U as Users
    participant S as Sessions
    participant L as Audit log
    B->>+G: POST /login
    G->>+A: verify credentials
    A->>U: find user
    U-->>A: user
    Note right of A: проверка пароля
    loop retry 3x
        A->>S: create session
        S-->>A: session id
    end
    A->>L: record login
    A-->>-G: token
    G-->>-B: 302 + cookie
```

## Links

See the [mermaid-ascii repository](https://github.com/AlexanderGrooff/mermaid-ascii) and https://example.com/docs.
