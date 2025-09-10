;;error:4:21-25:unknown symbol
;;error:4:31-32:expected u0, found u16
(defcolumns (X :i16))
(deflookup test ((+ m2.A 1)) (X))

(module m1)
(defcolumns (A :i16))
