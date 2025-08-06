;;error:3:41-42:recursion not permitted here
(defcolumns (X :i16))
(defcomputedcolumn (Y :i16) (+ X (shift Y -1)))
