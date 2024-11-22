(defcolumns ST A B)
(defconstraint c1 (:guard ST) (if A 0 B))
