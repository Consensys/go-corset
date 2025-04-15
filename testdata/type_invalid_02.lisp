;;error:7:26-29:expected bool, found u1
;;error:7:30-31:expected bool, found u8
(defcolumns
    (BIT :binary)
    (X :i8))

(defconstraint c1 () (if BIT X))
