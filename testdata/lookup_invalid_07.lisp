;;error:2:1-2:blah
(defcolumns X)
(deflookup test ((+ m1.A m2.B)) (X))

(module m1)
(defcolumns A)

(module m2)
(defcolumns B)
