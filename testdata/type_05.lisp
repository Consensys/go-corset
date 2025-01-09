(defcolumns
    (BIT :binary@loob)
    (X :i8@loob))

(defconstraint c1 () (if BIT X))
