(defcolumns ST A B C)
(defconstraint c1 (:guard ST) (if A B C))
