;;error:3:20-23:column is untyped
(defcolumns (X :i16))
(defcomputedcolumn (Y) (+ X 1))