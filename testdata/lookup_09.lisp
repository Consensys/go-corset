(defcolumns (X :i16))
(deflookup test (m1.A) (X))

(module m1)
(defcolumns (A :i16))
