(defcolumns ST A B)
(defconstraint c1 () (* ST (- 1 (~ A)) (- 1 (~ B))))
