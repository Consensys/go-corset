;;error:4:30-32:found 2 arguments, expected 1
(defcolumns (X :i16) (Y :i16))
(defpurefun (op x) (+ x 1))
(defconstraint c1 () (reduce op (begin X Y)))
