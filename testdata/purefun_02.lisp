(defcolumns A B)
(defpurefun (eq x y) (- y x))
(defconstraint test () (eq A B))
