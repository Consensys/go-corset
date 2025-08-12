(defcolumns (X :i16))
(defcomputedcolumn (Y :i16 :bwd) (+ X (shift Y 1)))
