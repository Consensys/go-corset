(defcolumns (X :@loob) (Y :@loob))
(defpurefun (op x) (+ x 1))
(defconstraint c1 () (reduce op (begin X Y)))
