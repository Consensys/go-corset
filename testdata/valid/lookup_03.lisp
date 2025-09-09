(defcolumns (X :i16) (Y :i32))
(deflookup test (Y) ((* X 2)))
