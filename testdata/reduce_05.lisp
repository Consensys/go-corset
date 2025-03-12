(defcolumns (X :i16@loob) (Y :i16@loob))
(defpurefun (op x y) (+ x y))
(defconstraint c1 () (reduce op (begin X Y)))
