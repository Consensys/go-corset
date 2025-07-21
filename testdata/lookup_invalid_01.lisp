;;error:3:21-26:differing number of source and target columns (2 v 1)
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y) (X X))
