(defcolumns A B)
(defpurefun ((eq :i16@loob) x y) (- y x))
(defconstraint test () (eq A B))
