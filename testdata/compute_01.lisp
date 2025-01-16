(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X)
(defcomputed (Y) (id X))
(defconstraint c1 () (vanishes! (- X Y)))
(defconstraint c2 () (vanishes! (- Y X)))
