(defcolumns ST A B)
;; STAMP == 0 || X == 1 || X == 2
(defconstraint c1 (:guard ST) (if A B))
