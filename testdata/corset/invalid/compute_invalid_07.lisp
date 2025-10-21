;;error:3:21-23:too many target columns (expected 1)
(defcolumns (X :i16))
(defcomputed (Y Z) (id X))
