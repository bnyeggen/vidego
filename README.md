Simple program to index movie files and maintain persistent metadata about them.  Really, it's about writing a first CRUD app in Go / Javascript.

TODO:

FRONTEND:
- Refactor Javascript code to isolate table configuration & have a somewhat proper MVC layout
- Infinite scroll / lazy rendering of table rows
- Switch to static, separately-stored dataset, rather than discarding once DOM is built
- Improve filter algorithm
  - Operate off of the static dataset rather than crawling DOM
  - Incrementally filter - if filter value is purely added to, restrict present subset
  - Benchmark replacing table body w/ new, filtered elements, rather than setting display attribute incrementally
- Accumulate all changes to table body and apply in one operation
- Autocomplete for director name

BACKEND:
- Switch from sqlite to an external DB (SQL or no)
- Conceptual concurrency (but probably parallelism of ~1 to avoid thrashing disk) for hashing & DB insertion
- Collection of distinct directors

GENERAL:
- Possibly implement tagging (eg for genre)
