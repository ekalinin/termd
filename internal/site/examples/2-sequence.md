## Sequence diagrams

Mermaid diagrams are drawn as text.

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant A as API
    C->>+A: GET /doc
    Note right of A: check token
    alt token is valid
        A-->>C: 200 OK
    else
        A-->>C: 401 Unauthorized
    end
    deactivate A
```
