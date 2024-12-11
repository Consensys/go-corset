(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y)
(defconstraint c1 () (vanishes! (* X Y)))
(defconstraint c2 () (vanishes! (* Y X)))
