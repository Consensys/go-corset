(defcolumns ST A)
(defconstraint spills () (* ST A (~ (shift A 3))))
