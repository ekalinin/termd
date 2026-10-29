# Tasks

## 1. Parsing and scheduling

- [x] 1.1 Add `TestGanttDates` to `internal/diagram/diagram_test.go`, checking the start, the end and the end of the bar of the parsed and scheduled tasks for the scenarios of "Gantt task dates": a task without a start, `after a b`, `until b` with `b` defined later, the durations `1w`, `1.5d`, `1M` from `2024-01-31` and `36h`, `dateFormat HH:mm` (`17:49` to `18:04`), `dateFormat X` (`0` to `71` seconds), `excludes weekends` (A ends at `2024-01-08` with its bar at `2024-01-06`, B from `2024-01-08` to `2024-01-09`), `includes 2024-01-06` (end `2024-01-08`), `inclusiveEndDates`, an end date that is not moved by `excludes`, the formats `DD.MM.YYYY` and `D/M/YY`, the rejected date `2024-02-30`, and a `vert` task that is the task before the next one; and the errors `unsupported line 3`, `invalid date on line 2`, `unsupported dateFormat`, `no start date on line 2`, `unknown task x`, `circular dependency` and `no tasks`; verify `go test ./internal/diagram -run TestGanttDates` fails
- [x] 1.2 Add `internal/diagram/gantt.go` with the parser, the `dateFormat` matcher, the durations, the scheduling in passes and the excluded days of design.md; verify the test of 1.1 passes

## 2. Layout

- [x] 2.1 Add `TestGantt` to `internal/diagram/diagram_test.go`, comparing the lines of `Render("mermaid", src, width)` for the scenarios of "Gantt layout": the bars at width 40 (the axis line, `Build`, `  Design` with 15 `█`, `  Code` with 15 `█` after 19 spaces), the title line and the frontmatter title (`Plan`, blank, the axis line; `Old` not printed), the task states (`░`, `▓`, `█` for `crit` and plain tasks, one `◆` in the last column), tasks before the first section, a `vert` task without a line, `axisFormat %-d %b` (`1 Jan`, `21 Jan`), an axis with only the start at width 30, and the ignored lines giving the same output as the chart without them; none of them falls back; and the frame labels `mermaid - gantt - <reason>` with `Fallback` set for the errors of 1.1; verify `go test ./internal/diagram -run 'TestGantt$'` fails
- [x] 2.2 Add the layout of design.md to `gantt.go` with the names column as wide as the widest name, and a `case "gantt"` in `Render`; verify the tests of 2.1 and 1.1 pass and `go test ./internal/diagram ./internal/render` passes with unchanged golden files

## 3. Width

- [x] 3.1 Add `TestGanttWidth` to `internal/diagram/diagram_test.go`: at width 40, a section with the task `Implement the parser and the tests :2024-01-01, 5d` gives a names column of 18 columns, the name wrapped with the bar on its first line, bars of 20 columns, no line wider than 40 and `Wide` unset; at width 25, the task `Implement the parser :2024-01-01, 5d` gives a names column of 10 columns, bars of 20 columns and `Wide` set; a word longer than the names column is split and every bar starts in the same column; a section name longer than the output wraps at the output width; verify `go test ./internal/diagram -run TestGanttWidth` fails
- [x] 3.2 Shrink the names column to `max(width - 22, 10)` with bars of at least 20 columns, wrap the names with `text.Wrap` splitting long words, and wrap the section names at the output width; verify the tests of 3.1, 2.1 and 1.1 pass

## 4. Golden fixtures

- [x] 4.1 Add `internal/diagram/testdata/gantt.mmd` with the syntax example of the Mermaid gantt documentation and `internal/diagram/testdata/gantt-basic.mmd` with its first example, and add both to `TestSupportedTypesAreRendered`; run `go test ./internal/diagram -run TestGolden -update`; verify `git status` shows only the two new golden files, the axis of `gantt` shows `2014-01-06` and `2014-01-28` and the bar of `Add to mermaid` ends in the column of the `◆` of `Functionality added`, and the axis of `gantt-basic` shows `2014-01-01` and `2014-02-20`
- [x] 4.2 Add `testdata/gantt.md` with the syntax example of the Mermaid documentation, a chart with a title and every task state, a chart with a task name long enough to wrap at 40 columns, and a chart with an unknown task followed by a paragraph; run `make golden`; verify `git status testdata/golden` shows only the six new `gantt.*.golden` files, the bars of a chart start in one column at every width, and the broken chart is a frame followed by the paragraph

## 5. Documentation

- [x] 5.1 In `README.md` add `gantt` to "Features" and to the "Diagrams" section (bars on a time axis under their sections, the start and the end of the chart on the axis, `░` for `done`, `▓` for `active`, `◆` for milestones, `crit` ignored, the dates as Mermaid computes them with `excludes`, long names wrapped, an unknown task shown as source), and remove `gantt` from "Limitations"; verify every statement matches `specs/diagram-rendering/spec.md`

## 6. Verification

- [x] 6.1 Run `make check` and `openspec validate render-gantt --strict`; verify both pass
- [x] 6.2 Run `make build`, then `./termd --no-pager --width 40 testdata/gantt.md` in a terminal and `./termd --width 80 testdata/gantt.md | cat`; verify the bars, the wrapped names, the frame of the broken chart, and no escape sequences in the pipe
