;;error:6:22-32:expected loobean constraint (found u1@bool)
(defcolumns
    (BIT :binary)
    (X :binary@bool))

(defconstraint c1 () (if BIT X))
