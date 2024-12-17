(defcolumns (X :@loob) (Y :@loob))
(defpurefun (op x y) (+ x y))
(defconstraint c1 () (reduce op (begin X Y)))
