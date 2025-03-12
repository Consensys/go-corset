;;error:3:23-26:incorrect number of columns
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y Y) (X))
