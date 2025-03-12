;;error:3:19-21:incorrect number of arguments (found 2)
(defcolumns (X :i16))
(defcomputed (Y) (id X X))
