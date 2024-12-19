;;error:3:33-40:incorrect number of arguments
(defcolumns ST A)
(defconstraint c1 () (* ST (- 1 (~ A 0))))
