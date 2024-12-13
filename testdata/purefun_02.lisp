(defcolumns A B)
(defpurefun ((eq :@loob) x y) (- y x))
(defconstraint test () (eq A B))
