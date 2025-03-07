;;error:3:18-19:unknown symbol
(defcolumns (X :i16))
(deflookup test (Y) (X))

(module m1)
(defcolumns Y)
