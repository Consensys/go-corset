;;error:3:19-21:found 2 arguments, expected 1
(defcolumns (X :i16))
(defcomputed (Y) (id X X))
