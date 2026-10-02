# Footnotes

A footnote reference shows the number of its note[^note]. Labels can be numbers[^1] or words, and the numbers follow the first reference, so the label `1` gets the number 2. A second reference to the same note[^note] shows the same number.

[^1]: A numeric label, defined in the middle of the document.

A reference in **bold text[^bold]** keeps the style of the text, and a table cell can hold one too:

| Feature | Notes |
|---|---|
| Footnotes | A reference in a cell[^cell] |

A reference without a definition[^missing] stays as written.

[^note]: This definition is long enough to wrap at every width of the golden files, and its wrapped lines start at the item text, not at the number.
[^bold]: The number is bold here.
[^cell]: A note with two paragraphs and a code block.

    The second paragraph is indented like the first one:

    ```sh
    termd --width 40 doc.md
    ```

[^unused]: This definition is never referenced and is not shown.
