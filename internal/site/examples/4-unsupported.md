## Unsupported diagrams

A diagram that cannot be drawn is shown as source in a labelled frame.

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Loading: open
    Loading --> Idle: done
```
