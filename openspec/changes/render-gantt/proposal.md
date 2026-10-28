# Proposal

## Why

A mermaid `gantt` block is shown as its source in a frame labelled "not supported". `gantt` is in the TODO list of the project and in "Limitations" of README. mermaid-ascii, which draws the other mermaid diagrams, has no gantt support at the commit termd uses, so termd draws gantt charts itself, as it does for mindmaps.

## What Changes

- termd draws a mermaid `gantt` as horizontal bars on a time axis: the task names on the left, grouped under their sections, and the bars on the right, scaled to the output width. A line above the bars shows the start and the end of the chart, formatted with `axisFormat`. A `title`, from the block or from its frontmatter, is printed above the chart.

  ```
  A Gantt Diagram

                     2014-01-01                     2014-02-20
  Section
    A task           █████████████████████████
    Another task                              ████████████████
  Another
    Task in Another           ██████████
    another task                        ████████████████████
  ```

- The dates of the tasks follow Mermaid: a start date, `after <id>...`, or the end of the task before it; an end date, a duration (`30d`, `24h`, `1w`, ...) or `until <id>...`; `dateFormat`, `excludes` with `weekends`, weekday names and dates, `includes`, `weekend` and `inclusiveEndDates`.
- The state of a task is shown by the fill of its bar, as diagrams have no styles: `░` for `done`, `▓` for `active`, `█` for the other tasks. A `milestone` is drawn as `◆`. `crit` is ignored.
- Task names that do not leave the bars 20 columns wrap in a column of at least 10 columns. A chart that is still wider than the output is emitted in full as a wide block.
- `click`, `todayMarker`, `tickInterval`, `weekday`, `topAxis`, `accTitle`, `accDescr`, `%%` comments and `displayMode` of the frontmatter are ignored, and a task with the `vert` tag is not drawn.
- A gantt chart that termd cannot read is shown as its source in a frame labelled with the reason, such as `unsupported line N`, `invalid date on line N` or `unknown task <id>`. Mermaid falls back to today's date for an unknown task in `after` or `until`; termd shows the source instead, so the output does not depend on the day it runs.

Out of scope:

- PlantUML gantt charts (`@startgantt`), which stay framed source labelled `plantuml - gantt - not supported`.
- Ticks between the start and the end of the axis, and `tickInterval`.
- Marking excluded days, the today marker, vertical markers (`vert`) and the compact display mode.
- Colors and the `crit` state.
- `dateFormat` tokens other than numbers: month names (`MMM`), quarters (`Q`), ordinals (`Do`), days of the year (`DDD`), 12-hour time (`h`, `a`), fractions of a second (`S`) and time zones (`Z`).
- An ASCII form of the bar characters; `add-ascii-mode` will need one for `█`, `▓`, `░` and `◆`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities
- `diagram-rendering`: "Supported mermaid types" adds `gantt`; "Framed source fallback" no longer names `gantt` as an unsupported mermaid type; new requirements "Gantt task dates" and "Gantt layout".

## Impact

- `internal/diagram`: a parser, a date scheduler and a layout for gantt charts in a new `gantt.go`, and a `gantt` case in `Render`; mermaid-ascii does not change.
- Tests and golden fixtures in `internal/diagram/testdata`, and a new `testdata/gantt.md` with its golden files; the existing golden files do not change.
- `README.md`: "Features", the "Diagrams" section and "Limitations", which list `gantt` as shown as source.
