;;error:6:22-32:expected loobean constraint (found u8)
(defcolumns
    (BIT :binary)
    (X :i8))

(defconstraint c1 () (if BIT X))
