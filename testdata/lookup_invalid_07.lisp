;;error:3:26-30:conflicting context
(defcolumns (X :i16))
(deflookup test ((+ m1.A m2.B)) (X))

(module m1)
(defcolumns A)

(module m2)
(defcolumns B)
