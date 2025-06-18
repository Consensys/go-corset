;;error:6:26-29:expected bool, found u1
(defcolumns
    (BIT :binary)
    (X :binary))

(defconstraint c1 () (if BIT (== 0 X)))
