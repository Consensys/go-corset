(defcolumns (A :i16 :padding 1))
(defconstraint c1 () (== 0 (- 1 (~ A))))
