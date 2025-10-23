;;error:3:25-29:unknown symbol
(defcolumns (X :i16))
(deflookup test (X) ((+ m2.A 1)))

(module m1)
(defcolumns (A :i16))
