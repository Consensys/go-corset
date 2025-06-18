(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16) (Z :i16))
(defconstraint c1 () (vanishes! (* Z (- Z 1))))
(defconstraint c2 () (vanishes! (* (- Y Z) (- Y Z 2))))
(defconstraint c3 () (vanishes! (* (- X Y) (- X Y 4))))
