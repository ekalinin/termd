# Spec Delta

## MODIFIED Requirements

### Requirement: Supported mermaid types
termd SHALL render mermaid `sequenceDiagram`, `flowchart`, `graph` and `erDiagram` blocks as text diagrams drawn with Unicode box-drawing characters, mermaid `mindmap` blocks as indented trees, and mermaid `gantt` blocks as bars on a time axis, in place of the code block.

#### Scenario: Flowchart
- **WHEN** a document contains a `mermaid` block starting with `flowchart LR`
- **THEN** the output shows the nodes as boxes connected by arrows, and the mermaid source is not shown

#### Scenario: Legacy graph keyword
- **WHEN** a `mermaid` block starts with `graph TD`
- **THEN** it is rendered as a flowchart

#### Scenario: Mindmap
- **WHEN** a document contains a `mermaid` block starting with `mindmap`
- **THEN** the output shows the nodes as an indented tree, and the mermaid source is not shown

#### Scenario: Gantt
- **WHEN** a document contains a `mermaid` block starting with `gantt`
- **THEN** the output shows the tasks as bars on a time axis, and the mermaid source is not shown

### Requirement: Diagram titles
termd SHALL print the `title` from a diagram's frontmatter above the rendered diagram. A gantt chart with a `title` line SHALL print the title of that line instead.

#### Scenario: Title from frontmatter
- **WHEN** a mermaid block has frontmatter `title: Login flow`
- **THEN** `Login flow` is printed above the diagram

#### Scenario: Gantt title line
- **WHEN** a `gantt` block has frontmatter `title: Old` and the line `title Plan`
- **THEN** `Plan` is printed above the chart, and `Old` is not printed

### Requirement: Framed source fallback
termd SHALL show the source of a diagram block inside a frame, instead of rendering it, when a PlantUML block is not a sequence diagram or a mindmap (for example a class diagram), when the mermaid type is not supported (for example `stateDiagram-v2`, `classDiagram`), or when the diagram cannot be parsed or rendered. The frame SHALL carry a label naming the diagram language, the detected type when known, and the reason. A failure in one diagram SHALL NOT stop the rest of the document from rendering.

#### Scenario: Unsupported mermaid type
- **WHEN** a document contains a `mermaid` block starting with `stateDiagram-v2`
- **THEN** the block's source is shown in a frame labelled with `mermaid`, `stateDiagram-v2` and "not supported", and the following content is rendered normally

#### Scenario: PlantUML block
- **WHEN** a document contains a `plantuml` block with the lines `@startuml`, `class User` and `@enduml`
- **THEN** its source is shown in a frame labelled with `plantuml`, `class` and "not supported"

#### Scenario: Invalid mermaid
- **WHEN** a `mermaid` block starts with `sequenceDiagram` but contains a syntax error
- **THEN** its source is shown in a frame labelled with `mermaid`, `sequenceDiagram` and the parse failure, and termd exits with status 0

#### Scenario: Renderer failure
- **WHEN** rendering a supported diagram fails unexpectedly
- **THEN** the diagram is shown as framed source and the rest of the document is rendered

## ADDED Requirements

### Requirement: Gantt task dates
termd SHALL read the lines of a mermaid `gantt` block after its `gantt` line as follows:

- `dateFormat F` sets the format of the dates of the block, `YYYY-MM-DD` by default. F SHALL be made of the tokens `YYYY`, `YY`, `MM`, `M`, `DD`, `D`, `HH`, `H`, `mm`, `m`, `ss` and `s`, and of other characters that stand for themselves, or F SHALL be `X`, seconds since 1970-01-01, or `x`, milliseconds since 1970-01-01. `M`, `D`, `H`, `m` and `s` accept one or two digits, the other tokens their full number of digits, and `YY` is a year from 1969 to 2068. The parts of a date that F does not have SHALL be those of 1970-01-01 00:00:00. Dates SHALL be read and shown in UTC, so the output does not depend on the time zone of the machine.
- `title text`, `section name` and `axisFormat` are used by "Gantt layout".
- `excludes` and `includes` name days by tokens separated by spaces or commas, in any case: `weekends`, weekday names such as `sunday`, and dates written in the `dateFormat` or as `YYYY-MM-DD`. Several `excludes` and `includes` lines add up. `weekend friday` makes the weekend Friday and Saturday, and `weekend saturday` Saturday and Sunday, the default.
- `inclusiveEndDates` makes every end given as a date one day later.
- `click`, `todayMarker`, `tickInterval`, `weekday`, `topAxis`, `accTitle:`, `accDescr:`, `accDescr {` ... `}`, empty lines and lines that start with `%%` are ignored.
- Any other line with a `:` is a task: the text before the first `:` is its name, and the text after it is a list of items separated by commas. The items SHALL start with any of the tags `done`, `active`, `crit`, `milestone` and `vert`, followed by one, two or three items: an end; a start and an end; or an id, a start and an end.

The start of a task SHALL be a date, or `after` followed by ids separated by spaces, which is the latest end of these tasks. A task without a start SHALL start at the end of the task before it in the block, in any section. The end of a task SHALL be a date; `until` followed by ids, which is the earliest start of these tasks; or a duration, a number with an optional fractional part followed by `ms`, `s`, `m`, `h`, `d`, `w`, `M` or `y`, added to the start. `d` and `w` SHALL add days, the number of days rounded to the nearest whole number; `M` and `y` SHALL add calendar months and years without the fractional part, ending on the last day of the month when the month has fewer days; the other units SHALL add exact time. An end that is not a date, an `until` or a duration SHALL be the start of the task. An id SHALL name a task before or after the task that uses it.

A day SHALL be excluded when `excludes` names its date or its weekday, or has `weekends` and the day is a day of the weekend, unless `includes` names its date. When the end of a task is a duration and a day is excluded, termd SHALL check the days after the start one by one up to the end, and every excluded day SHALL move the end one day later. The tasks that start at the end of this task SHALL start at the moved end, and its bar SHALL end before the excluded days right before the moved end.

termd SHALL show the block as framed source labelled with `mermaid`, `gantt` and the reason:

- `unsupported line N` for any other line, a task without items after its tags, and a task with more than three of them;
- `invalid date on line N` for a start that is not a date in the `dateFormat` and not an `after`;
- `unsupported dateFormat` for a `dateFormat` with other tokens of Mermaid, such as `MMM`, `Q`, `Do`, `DDD`, `hh`, `A`, `SSS` or `Z`;
- `no start date on line N` for a task without a start and without a task before it;
- `unknown task` followed by the id for an `after` or an `until` with an id that no task has;
- `circular dependency` for tasks whose starts or ends depend on each other;
- `no tasks` for a block without tasks, not counting the tasks with the `vert` tag.

N SHALL be the number of the line in the block, starting with 1.

#### Scenario: Task without a start
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, 10d` and `B :10d`
- **THEN** the axis shows `2024-01-01` and `2024-01-21`, and the bar of B starts in the column where the bar of A ends

#### Scenario: After several tasks
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, 5d`, `B :b, 2024-01-01, 10d` and `C :after a b, 10d`
- **THEN** the axis shows `2024-01-01` and `2024-01-21`, and the bar of C starts in the column where the bar of B ends

#### Scenario: Until a later task
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, until b` and `B :b, 2024-01-11, 10d`
- **THEN** the axis shows `2024-01-01` and `2024-01-21`, and the bar of A ends in the column where the bar of B starts

#### Scenario: Durations
- **WHEN** gantt charts with `dateFormat YYYY-MM-DD` have the single tasks `A :2024-01-01, 1w`, `A :2024-01-01, 1.5d`, `A :2024-01-31, 1M` and `A :2024-01-01, 36h`
- **THEN** their axes end at `2024-01-08`, `2024-01-03`, `2024-02-29` and `2024-01-02`

#### Scenario: Times
- **WHEN** a gantt chart has the lines `dateFormat HH:mm`, `axisFormat %H:%M`, `A :a, 17:49, 10m` and `B :5m`
- **THEN** the axis shows `17:49` and `18:04`

#### Scenario: Unix timestamps
- **WHEN** a gantt chart has the lines `dateFormat X`, `axisFormat %s` and `A :0, 71`
- **THEN** the axis shows `0` and `71`

#### Scenario: Excluded weekend
- **WHEN** a gantt chart has the lines `excludes weekends`, `A :a, 2024-01-01, 5d` and `B :after a, 1d`
- **THEN** the axis shows `2024-01-01` and `2024-01-09`, the bar of A ends in the column of `2024-01-06`, the bar of B starts in the column of `2024-01-08`, and the columns between them are empty

#### Scenario: Included day
- **WHEN** a gantt chart has the lines `excludes weekends`, `includes 2024-01-06` and `A :2024-01-01, 6d`
- **THEN** the axis ends at `2024-01-08`

#### Scenario: Inclusive end dates
- **WHEN** a gantt chart has the lines `inclusiveEndDates` and `A :2024-01-01, 2024-01-10`
- **THEN** the axis ends at `2024-01-11`

#### Scenario: Ignored lines
- **WHEN** a gantt chart has the lines `todayMarker off`, `tickInterval 1week`, `weekday monday`, `topAxis`, `accTitle: Plan`, `click a href "https://example.com"` and `%% a comment` besides its tasks
- **THEN** the output is the same as for the chart without these lines

#### Scenario: Unsupported line
- **WHEN** a `mermaid` block has the lines `gantt`, `dateFormat YYYY-MM-DD` and `A 2024-01-01, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `unsupported line 3`

#### Scenario: Invalid date
- **WHEN** a `mermaid` block has the lines `gantt` and `A :2024-13-01, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `invalid date on line 2`

#### Scenario: Unsupported date format
- **WHEN** a `mermaid` block has the lines `gantt`, `dateFormat MMM D YYYY` and `A :Jan 1 2024, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `unsupported dateFormat`

#### Scenario: No start date
- **WHEN** a `mermaid` block has the lines `gantt` and `A :3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `no start date on line 2`

#### Scenario: Unknown task
- **WHEN** a gantt chart has the task `A :after x, 3d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `unknown task x`

#### Scenario: Circular dependency
- **WHEN** a gantt chart has the tasks `A :a, after b, 1d` and `B :b, after a, 1d`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `circular dependency`

#### Scenario: No tasks
- **WHEN** a `mermaid` block has the lines `gantt` and `title Plan`
- **THEN** its source is shown in a frame labelled with `mermaid`, `gantt` and `no tasks`

### Requirement: Gantt layout
termd SHALL draw a mermaid `gantt` block as lines without styles:

- The title, from the `title` line or else from the frontmatter, followed by a blank line.
- The axis line: the start of the chart, the earliest start of a task, at the first column of the bars, and the end of the chart, the latest end of a task, ending at the last column of the bars. Both SHALL be formatted with `axisFormat`, `%Y-%m-%d` by default. When the bars have fewer columns than the two dates and a space between them, only the start SHALL be shown.
- Every section with tasks: its name on its own line, wrapped at the output width, followed by its tasks in the order of the block. The tasks before the first section SHALL come first, without a section line.
- Every task on its own line: its name, indented by 2 spaces in a section, padded to the width of the names column, 2 spaces, and its bar. The names column SHALL be as wide as the widest name with its indentation.

The bars SHALL take the columns from the end of the names column and its 2 spaces up to the output width. They SHALL map the time from the start to the end of the chart evenly onto their columns: the instant t SHALL fall at column round(C × (t − start) / (end − start)), counting from 0, where C is the number of columns of the bars, and at column 0 when the start and the end of the chart are the same. A bar SHALL fill the columns from the column of the start of its task up to, not including, the column of the end of its bar, and at least one column; a bar that starts at column C SHALL fill column C − 1. A bar SHALL be made of `▓` for an `active` task, `░` for a `done` task that is not `active`, and `█` for any other task. A `milestone` SHALL be a single `◆` at the column of the middle between its start and its end, and at column C − 1 when that column is C. A task with the `vert` tag SHALL NOT be drawn and SHALL NOT take a line. The `crit` tag SHALL NOT change the bar.

When the bars would have fewer than 20 columns, the names column SHALL shrink to leave them 20 columns, but SHALL NOT be narrower than 10 columns, and the bars SHALL have at least 20 columns. A name wider than its column SHALL wrap at whitespace, a word wider than the column SHALL be split, and the continuation lines SHALL start at the column of the name and have no bar. A chart that is still wider than the output SHALL be emitted in full as a wide block.

`axisFormat` SHALL support `%Y`, `%y`, `%m`, `%d`, `%e`, `%H`, `%I`, `%M`, `%S`, `%L`, `%p`, `%a`, `%A`, `%b`, `%B`, `%j`, `%s` and `%%` as in Mermaid, with English names of months and weekdays, and a `-` after the `%` SHALL drop the padding of a number. Any other `%` code SHALL be shown as written.

#### Scenario: Bars
- **WHEN** a gantt chart with the lines `dateFormat YYYY-MM-DD`, `section Build`, `Design :d, 2024-01-01, 10d` and `Code :after d, 10d` is rendered at width 40
- **THEN** the output is the axis line with `2024-01-01` at columns 11 to 20 and `2024-01-21` at columns 31 to 40, the line `Build`, the line `  Design` followed by 2 spaces and 15 `█`, and the line `  Code` followed by 19 spaces and 15 `█`

#### Scenario: Title
- **WHEN** a gantt chart has the line `title Plan`
- **THEN** the output starts with `Plan`, a blank line and the axis line

#### Scenario: Task states
- **WHEN** a gantt chart has the tasks `Design :done, d, 2024-01-01, 10d`, `Build :active, b, after d, 10d`, `Test :t, after b, 10d`, `Fix :crit, f, after t, 10d` and `Release :milestone, after f, 0d`
- **THEN** the bar of Design is made of `░`, the bar of Build of `▓`, the bars of Test and Fix of `█`, and the line of Release has a single `◆` in the last column of the bars

#### Scenario: Tasks without a section
- **WHEN** a gantt chart has the task `Kickoff :k, 2024-01-01, 1d`, then `section Work` and `Build :after k, 5d`
- **THEN** the line of Kickoff comes first and starts with `Kickoff`, followed by the line `Work` and the line of Build, which starts with `  Build`

#### Scenario: Vertical marker
- **WHEN** a gantt chart has the tasks `A :a, 2024-01-01, 10d` and `Deadline :vert, 2024-01-05, 0d`
- **THEN** the output has no line for Deadline

#### Scenario: Long names
- **WHEN** a gantt chart has a section with the task `Implement the parser and the tests :2024-01-01, 5d` and is rendered at width 40
- **THEN** the names column is 18 columns wide, the name wraps with the bar on its first line, the bars take 20 columns, and no line is wider than 40 columns

#### Scenario: Narrow output
- **WHEN** a gantt chart with a section and the task `Implement the parser :2024-01-01, 5d` is rendered at width 25
- **THEN** the names column is 10 columns wide, the bars take 20 columns, and the chart is emitted as a wide block

#### Scenario: Axis format
- **WHEN** a gantt chart with the line `axisFormat %-d %b` has the task `A :2024-01-01, 20d`
- **THEN** the axis shows `1 Jan` and `21 Jan`

#### Scenario: Narrow bars
- **WHEN** a gantt chart with the line `axisFormat %Y-%m-%dT%H:%M` and the task `A :2024-01-01, 5d` is rendered at width 30
- **THEN** the axis line shows only the start of the chart
