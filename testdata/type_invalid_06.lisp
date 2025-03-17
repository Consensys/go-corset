;;error:4:27-30:unexpected loobean guard
(defcolumns (BIT :i1) (X :i1))

(defconstraint c1 (:guard BIT) X)
