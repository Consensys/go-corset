;;error:4:22-45:incorrect number of arguments (expected 2)
(defcolumns (X :i16@loob) (Y :i16@loob))
(defpurefun (op x y z) (+ x y z))
(defconstraint c1 () (reduce op (begin X Y)))
