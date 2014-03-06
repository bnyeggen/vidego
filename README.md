Simple program to index movie files and maintain persistent metadata about them.  Really, it's about writing a first CRUD app in Go / Javascript.

TODO:

FRONTEND:
- Autocomplete for director name

BACKEND:
- Switch from sqlite to an external DB (SQL or no)
- Conceptual concurrency (but probably parallelism of ~1 to avoid thrashing disk) for hashing & DB insertion
- Collection of distinct directors

GENERAL:
- Possibly implement tagging (eg for genre)
