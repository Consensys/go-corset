(defpurefun (vanishes! x) (== 0 x))

(defcolumns (X :i16) (Y :i16))
(defconstraint c1 () (vanishes! (* X Y)))
(defconstraint c2 () (vanishes! (* Y X)))
