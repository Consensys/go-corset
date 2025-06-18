;;error:7:26-29:expected bool, found u8
;;error:7:30-31:expected bool, found u8
(defcolumns
    (BIT :i8)
    (X :i8))

(defconstraint c1 () (if BIT X))
