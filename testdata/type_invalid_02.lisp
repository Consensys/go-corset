;;error:6:26-29:invalid condition (neither loobean nor boolean)
(defcolumns
    (BIT :binary)
    (X :i8@loob))

(defconstraint c1 () (if BIT X))
