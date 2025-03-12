(defcolumns (ST :i4) (X :i16) (Y :i16))
(deflookup test (Y) ((* ST (+ 1 X))))
