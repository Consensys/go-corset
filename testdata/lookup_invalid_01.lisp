;;error:3:21-26:incorrect number of columns
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y) (X X))
