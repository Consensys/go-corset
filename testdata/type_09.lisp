(defcolumns (BIT :i1@prove) (X :i4))
(defconstraint c1 () (if (!= 0 BIT) (== 0 X)))
