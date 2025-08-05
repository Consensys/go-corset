;;error:3:33-37:conflicting direction of recursion
(defcolumns (X :i16))
(defcomputedcolumn (Y :i16 :fwd :bwd) (+ X (shift Y -1)))
