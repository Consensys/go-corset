(defcolumns (X :i16))

(defcomputedcolumn (Y) (+ X 1))

(defconstraint c () eq! Y (+ X 1))