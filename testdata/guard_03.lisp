(defcolumns ST (A :@loob) B)
(defconstraint c1 (:guard ST) (if A 0 B))
