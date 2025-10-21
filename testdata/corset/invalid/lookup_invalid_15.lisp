;;error:3:22-23:expected u16, found u17
(defcolumns (X :i17) (Y :i16))
(deflookup test (Y) (X))
