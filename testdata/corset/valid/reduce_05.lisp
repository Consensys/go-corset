(defcolumns (X :i16) (Y :i16))
(defpurefun (op x y) (+ x y))
(defconstraint c1 () (== 0 (reduce op (begin X Y))))
