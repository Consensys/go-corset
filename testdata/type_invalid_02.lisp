;;error:2:1-2:blah
(defcolumns
    (BIT :binary)
    (X :i8@loob))

(defconstraint c1 () (if BIT X))
