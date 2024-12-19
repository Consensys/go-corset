;;error:4:21-25:unknown symbol
;;error:4:26-30:unknown symbol
(defcolumns X)
(deflookup test ((+ m1.A m2.B)) (X))

(module m1)
(defcolumns A)

(module m2)
(defcolumns B)
