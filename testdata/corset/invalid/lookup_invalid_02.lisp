;;error:3:23-26:differing number of source and target columns (1 v 2)
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y Y) (X))
