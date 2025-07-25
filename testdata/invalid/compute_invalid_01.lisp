;;error:3:20-21:malformed target columns
(defcolumns (X :i16))
(defcomputed Y (id X))
