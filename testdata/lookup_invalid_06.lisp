;;error:4:18-22:unknown symbol
;;error:4:23-27:unknown symbol
(defcolumns X Y)
(deflookup test (m1.A m2.B) (X Y))

(module m1)
(defcolumns A)

(module m2)
(defcolumns B)
