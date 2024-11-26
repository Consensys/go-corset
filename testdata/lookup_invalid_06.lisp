(defcolumns X Y)
(deflookup test (m1.A m2.B) (X Y))

(module m1)
(defcolumns A)

(module m2)
(defcolumns B)
