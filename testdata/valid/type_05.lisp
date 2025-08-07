(defcolumns
    (BIT :binary)
    (X :i8))

(defconstraint c1 () (if (== 0 BIT) (== 0 X)))
