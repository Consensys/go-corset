;;error:4:27-30:unexpected loobean guard
(defcolumns (BIT :i1@loob) (X :i1@loob))

(defconstraint c1 (:guard BIT) X)
