# Spec Delta

## MODIFIED Requirements

### Requirement: Diagram fallbacks do not fail the run
termd SHALL exit with status 0 when the document was rendered, even if some diagrams were shown as source because they are unsupported or invalid.

#### Scenario: Document with an unsupported diagram
- **WHEN** termd renders a document containing a `classDiagram` block
- **THEN** the diagram is shown as framed source and termd exits with status 0
