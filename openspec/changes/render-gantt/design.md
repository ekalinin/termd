# Design

## Context

`diagram.Render` passes the mermaid types `sequenceDiagram`, `erDiagram`, `flowchart` and `graph` to mermaid-ascii and draws `mindmap` itself with `renderMindmap` in `internal/diagram/mindmap.go`; any other type is framed source labelled `mermaid - <type> - not supported`. An error of a renderer goes through `reason` into the frame label, and `result` turns the drawn text into plain lines, wide when the widest line is wider than the output. `headerLine` finds the type line after the frontmatter and `%%` comments, `frontmatterTitle` reads the frontmatter `title`, and `hasKeyword` in `plantuml.go` matches a keyword alone or followed by a space. `internal/text` has `Width` and `Wrap`, which can split long words.

mermaid-ascii has no gantt support. The scheduling rules come from the Mermaid sources (`gantt.jison` and `ganttDb.js` on the `develop` branch) and the gantt page of its documentation:

- Keywords such as `dateFormat` and `section` are matched before tasks, case-sensitively; any other line with a `:` is a task.
- Dates are parsed with dayjs in strict mode, in the local time zone. A date without a year, a month or a day takes them from the current date.
- `after` and `until` with an unknown id fall back to today's date. A first task without a start crashes the parser.
- Days and weeks are added as `Math.round(n)` and `Math.round(7 × n)` days, months and years without the fractional part, clamped to the last day of the month.
- With `excludes`, `checkTaskDates` walks the days from the day after the start up to the end and moves the end one day for every excluded day; `renderEndTime`, the end of the bar, stays before the excluded days that end the task. It skips tasks whose end is a literal `YYYY-MM-DD` date, but not tasks whose end is `until` or a date in another format.
- A `vert` task takes no row but is still the task before the next one.

The markdown files of the Go module cache have no gantt block. The requirements are in `specs/diagram-rendering/spec.md`.

## Goals / Non-Goals

**Goals:**
- The dates of the tasks are the dates Mermaid computes for the charts of its documentation.
- The output depends only on the block and the width, not on the day or the time zone, so golden files are stable.
- The other diagram types render byte for byte as before; mermaid-ascii does not change.

**Non-Goals:**
- Parity with every quirk of Mermaid, listed under Decisions where termd differs.
- Configuration from the frontmatter, such as `config.gantt.axisFormat` or `displayMode`.

## Decisions

### A parser, a scheduler and a layout of termd in `internal/diagram/gantt.go`

`renderGantt(src string, width int) (string, error)` parses the block into a list of tasks and sections, schedules the tasks, and lays them out; `Render` gets a `case "gantt"` that calls it, so plain lines, `Wide` and `fallback` with `reason` are shared with the other types. The errors are plain `errors.New` and `fmt.Errorf` values with the reasons of the spec, such as `unsupported line 3`; `reason` keeps them as they are, because they do not start with `failed to`.

Alternatives:
- A table with the start, the end and the duration of every task. Rejected with the user in favor of bars.
- Translating the chart into a flowchart for mermaid-ascii. Rejected: a flowchart has no time axis.

### Parsing

The parser walks the lines after `headerLine`, keeping their index in the block for `line N`. It matches the keywords of the spec first, with `hasKeyword`, and with `accTitle:`, `accDescr:` and `accDescr {` by prefix; `accDescr {` skips the lines up to the one that contains `}`. Then a line with a `:` is a task: the name is the trimmed text before the first `:`, the items are the trimmed parts of the rest split at `,`. Leading items that are tags are removed; the remaining one, two or three items are the end, the start and the end, or the id, the start and the end. A task without an id has none, so `after task1`, which refers to an automatic id of Mermaid, is an unknown task. When two tasks have the same id, the later one wins, as in `taskDb` of Mermaid.

A task without a start records the index of the task before it in the block, `vert` tasks included; the first task without a start is the error `no start date on line N`. The title is the `title` line, or `frontmatterTitle(lines[:h])` without one.

### Dates without the time package layouts

`dateFormat` is split into tokens, longest first (`YYYY` before `YY`, `MM` before `M`, `DD` before `D`, `HH` before `H`, `mm` before `m`, `ss` before `s`), and literal characters. Tokens of dayjs that termd does not read (`MMMM`, `MMM`, `Q`, `Do`, `DDDD`, `DDD`, `hh`, `h`, `A`, `a`, `SSS`, `SS`, `S`, `ZZ`, `Z`, and `X` or `x` with other characters) are the error `unsupported dateFormat`. A date is matched against the tokens: a padded token reads its exact number of digits, a short one one or two, a literal must be equal. The fields go into `time.Date` in UTC with the missing parts from 1970-01-01, and the date is invalid unless `time.Date` kept every field, so `2024-02-30` is rejected as dayjs in strict mode rejects it. `X` reads a number of seconds with an optional fraction and `x` an integer number of milliseconds.

Alternative: translating `dateFormat` into a layout for `time.Parse`. Rejected: a layout reserves literal digits and words such as `1`, `Jan` or `PM`, and `time.Parse` accepts one or two digits for `15`, so a format with such literals would need escaping that the layout syntax does not have.

`axisFormat` is formatted by a small function over the codes of the spec, with English names; an unknown code is copied as written.

### Durations

`ms`, `s`, `m` and `h` are added as a `time.Duration` of the number times the unit. `d` and `w` are added with `AddDate` as `math.Round(n)` and `math.Round(7 × n)` days. `M` and `y` are added by a helper that moves the year and the month by the whole part of the number and clamps the day to the length of the target month, because `AddDate` normalizes January 31 plus one month to March 2 while Mermaid gives February 29.

### Scheduling in passes

Every task has a start and an end, each unresolved at first. A pass goes through the tasks in block order: a start is resolved when it is a date, when the tasks of its `after` have resolved ends (the latest of them), or when the task before it has a resolved end; an end is resolved when the start is resolved and the end is a date, a duration or an invalid end, or when the tasks of its `until` have resolved starts (the earliest of them). Passes repeat while one of them resolves something. Before the first pass, every id of an `after` and an `until` is looked up, and the first unknown one in block order is the error `unknown task <id>`; tasks still unresolved when a pass resolves nothing are the error `circular dependency`.

Mermaid stops after 10 passes and leaves the rest of the tasks without dates; termd repeats until nothing changes, which ends after at most two passes per task, one for its start and one for its end.

### Excluded days

`excludes` and `includes` tokens are lowercased and split at spaces and commas. A token is `weekends`, a weekday name, or a date read with the `dateFormat` or with `YYYY-MM-DD`; the dates are kept as calendar days, and other tokens are ignored, as they never match in Mermaid. A day is excluded by its calendar day and its weekday in UTC.

When a task with a duration end is resolved and a day is excluded, termd walks the days from the start plus one day while the day is not after the end, as `fixTaskDates` does: every excluded day moves the end one day later, and the end of the bar is the end as it was before the excluded days right before it. The end of the bar is only used for drawing; `after`, the task after it and the end of the chart use the moved end. When the end moves more than 10000 days, as it does when every weekday is excluded, the chart is framed source labelled `every day is excluded`; Mermaid throws an error after the same number of days.

termd moves only the ends that are durations. Mermaid also moves an `until` end and an end date in a format other than `YYYY-MM-DD`, so a task `until` a milestone overlaps it when a weekend lies between them. Not moving them keeps the `until` task ending where the milestone starts, which is what the line says.

### Layout

The layout collects the drawn rows: the tasks before the first section, then every section with its tasks, without `vert` tasks. The chart starts at the earliest start and ends at the latest end of the drawn tasks; an end before its start is drawn as the start.

The names column is the widest name plus its indentation, measured with `text.Width`. When the output leaves fewer than 20 columns for the bars, the column becomes `max(width - 22, 10)` and the bars `max(width - column - 2, 20)`. Names are wrapped with `text.Wrap` at the column minus the indentation, splitting long words, so every bar starts at the same column. Section names are wrapped at the output width without splitting words.

A column is `round(C × (t − start) / (end − start))` computed with `float64` nanoseconds; adjacent tasks therefore meet at one column without a gap or an overlap. Every line is trimmed on the right, as the lines of mindmaps are, and `result` marks the block wide.

### Fixtures

`internal/diagram/testdata/gantt.mmd` holds the syntax example of the Mermaid documentation (`excludes weekends`, `done`, `active`, `crit`, `until`, a milestone, four sections) and `gantt-basic.mmd` the first example of the page; both go into `TestSupportedTypesAreRendered` and `TestGolden`. Scheduling is tested on the parsed tasks, which is where the dates are exact; the layout is tested on the lines of `Render`.

## Risks / Trade-offs

- [`█`, `▓`, `░` and `◆` are East Asian ambiguous characters, which some CJK terminals draw two columns wide] → The box-drawing characters of the other diagrams have the same width class, so all diagrams are affected alike; `add-ascii-mode` will need ASCII forms for them.
- [Only the start and the end of the chart are on the axis, so the date of a bar in the middle has to be estimated] → Chosen with the user; ticks can be added later without changing the scheduling.
- [An end that is neither a date nor a duration, such as `3dX`, silently becomes a one-column bar] → Mermaid does the same and only logs a warning; the bar is still drawn at its start.
- [Dates in UTC differ from Mermaid for a block that depends on the local time zone, for example a `dateFormat X` chart near midnight with `excludes`] → Mermaid's output for such a chart already differs between machines; termd gives one answer everywhere.
- [No gantt block in the module cache, so there is no real-world sample] → The fixtures use the examples of the Mermaid documentation.

## Migration Plan

None. Gantt charts were framed source, and no existing fixture or golden file has one. Rollback is a revert of the change.
