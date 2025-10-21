(defcolumns (A :i16) (B :i16))
(defconstraint c1 () (== 0 (~ (- A B))))
