(defpurefun (eq! x y) (== x y))

(defcolumns (X :i16))

(defcomputedcolumn (Y :i24) (+ X 1))

(defconstraint c () (eq! Y X ))
