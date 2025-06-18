;;error:4:30-32:found 2 arguments, expected 3
(defcolumns (X :i16) (Y :i16))
(defpurefun (op x y z) (+ x y z))
(defconstraint c1 () (reduce op (begin X Y)))
