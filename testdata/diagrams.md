# Diagrams

```mermaid
sequenceDiagram
    participant C as Client
    participant A as API
    C->>+A: GET /doc
    Note right of A: проверка токена
    A-->>-C: 200 OK
```

```mermaid
classDiagram
    class User
    User --> Session: opens
```

The paragraph after the diagrams is rendered normally.
