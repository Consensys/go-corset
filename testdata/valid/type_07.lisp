(defcolumns
    (BIT :binary)
    (X :i8))

(defconstraint c1 () (if (== 0 (+ 1 BIT)) (== 0 X)))
