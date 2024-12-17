(defcolumns (X :@loob) (Y :@loob))
(defpurefun (op x y z) (+ x y z))
(defconstraint c1 () (reduce op (begin X Y)))
