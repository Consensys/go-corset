(defcolumns (ST :i16) (A :i16))
(defconstraint c1 () (== 0 (* ST (- 1 (~ A)))))
