(defcolumns
    (BIT :binary@loob)
    (X :binary@bool))

(defconstraint c1 () (if BIT X))
