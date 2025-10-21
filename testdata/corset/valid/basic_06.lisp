(defcolumns (X :i16))
(defconstraint c1 () (== 0 (* X (- X 1))))
(defconstraint c2 () (∨ (== 0 X) (== X 1)))
