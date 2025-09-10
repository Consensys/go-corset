;;error:3:22-29:expected u16, found u17
(defcolumns (X :i16) (Y :i16))
(deflookup test (Y) ((* 2 X)))
