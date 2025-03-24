(module m1)

(defcolumns
    (ACC_1 :i128)
    (BYTE :byte :array [0:2])
)

(defconstraint c1 () (== 0 (if (== ACC_1 1) [BYTE 0])))
