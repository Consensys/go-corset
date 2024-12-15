;;error:2:1-2:blah
(defcolumns
    (BIT :binary@loob)
    (X :i8))

(defconstraint c1 () (if BIT X))
