## Tables

Widths are measured by grapheme clusters, so emoji and CJK text keep columns aligned. Only the long column wraps, and a link shows its text instead of the URL.

| Sample | Cols | Notes | Spec |
|:---:|---:|---|---|
| 日本語 | 6 | CJK characters take two columns each | [UAX #11](https://www.unicode.org/reports/tr11/) |
| 👩‍💻 | 2 | A ZWJ sequence is one grapheme cluster | [UAX #29](https://www.unicode.org/reports/tr29/) |
| ⚠️ | 2 | Terminals disagree on emoji with a variation selector; termd follows Unicode and counts 2 | [UTS #51](https://www.unicode.org/reports/tr51/) |
