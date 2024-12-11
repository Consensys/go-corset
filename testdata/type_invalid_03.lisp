(defcolumns
    (BIT :i8)
    (X :i8@loob))

(defconstraint c1 () (if BIT X))
