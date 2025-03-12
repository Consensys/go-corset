;;error:3:22-23:unknown symbol
(defcolumns (X :i16))
(deflookup test (X) (Y))

(module m1)
(defcolumns (Y :i16))
