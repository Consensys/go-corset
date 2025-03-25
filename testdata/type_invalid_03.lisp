;;error:6:26-29:expected bool, found u8
(defcolumns
    (BIT :i8)
    (X :i8))

(defconstraint c1 () (if BIT X))
