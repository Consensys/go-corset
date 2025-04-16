;;error:6:26-35:expected bool, found int
(defcolumns
    (BIT :binary)
    (X :binary))

(defconstraint c1 () (if (+ 2 BIT) (== 0 X)))
