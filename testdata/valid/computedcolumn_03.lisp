(defcolumns (X :i16))
(defcomputedcolumn (Y :i16 :fwd) (+ X (shift Y -1)))
