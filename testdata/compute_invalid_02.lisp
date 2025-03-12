;;error:3:14-16:malformed target columns
(defcolumns (X :i16))
(defcomputed () (id X))
