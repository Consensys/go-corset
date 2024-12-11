(defcolumns
    (BIT :binary@bool)
    (X :i8))

(defconstraint c1 () (if BIT X))
