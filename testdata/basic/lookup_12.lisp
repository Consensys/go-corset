(defconst ONE 1)
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y) ((* ONE X)))
