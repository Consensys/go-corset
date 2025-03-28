(defcolumns (BIT :i1) (X :i1))

(defconstraint c1 (:guard BIT) (== 0 X))
