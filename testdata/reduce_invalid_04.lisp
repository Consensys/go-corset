;;error:4:30-32:unknown symbol
(defcolumns (X :i16) (Y :i16))
(defpurefun (op x y z) (+ x y z))
(defconstraint c1 () (reduce op (begin X Y)))
