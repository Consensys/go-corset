(defcolumns ST A B)
(defconstraint spills () (* ST A (~ (* (shift A 3) (shift B 2)))))
