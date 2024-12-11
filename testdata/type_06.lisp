(defcolumns
    (BIT :binary@bool)
    (X :i8@loob))

(defconstraint c1 () (if BIT X))
