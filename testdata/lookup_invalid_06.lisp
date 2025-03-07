;;error:3:23-27:conflicting context
(defcolumns (X :i16) (Y :i16))
(deflookup test (m1.A m2.B) (X Y))

(module m1)
(defcolumns A)

(module m2)
(defcolumns B)
