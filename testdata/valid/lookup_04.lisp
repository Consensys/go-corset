(defcolumns (ST :i4) (X :i16) (Y :i20))
(deflookup test (Y) ((* ST (+ 1 X))))
