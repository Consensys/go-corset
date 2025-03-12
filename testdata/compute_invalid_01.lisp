;;error:3:14-15:malformed target columns
(defcolumns (X :i16))
(defcomputed Y (id X))
