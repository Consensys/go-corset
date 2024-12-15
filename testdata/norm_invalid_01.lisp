;;error:2:1-2:blah
(defcolumns ST A)
(defconstraint c1 () (* ST (- 1 (~ A 0))))
