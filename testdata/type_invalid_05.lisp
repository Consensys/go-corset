(defcolumns
    (BIT :binary@loob)
    (X :binary@loob))

(defconstraint c1 () (if (+ 2 BIT) X))
