;;error:6:26-35:invalid condition (neither loobean nor boolean)
(defcolumns
    (BIT :binary)
    (X :binary@loob))

(defconstraint c1 () (if (+ 2 BIT) X))
