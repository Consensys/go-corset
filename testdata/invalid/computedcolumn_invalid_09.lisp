;;error:3:33-37:not declared recursive
(defcolumns (X :i16))
(defcomputedcolumn (Y :i16) (+ X (shift Y -1)))
