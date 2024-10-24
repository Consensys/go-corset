(defcolumns ST A B)
(defconstraint c1 (:guard ST) (ifnot A B))
