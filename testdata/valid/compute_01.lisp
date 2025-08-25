(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16))
(defcomputed ((Y :i16)) (id X))
(defconstraint c1 () (vanishes! (- X Y)))
(defconstraint c2 () (vanishes! (- Y X)))
