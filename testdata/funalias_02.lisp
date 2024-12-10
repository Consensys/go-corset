(defcolumns A B)
(defpurefun (eq x y) (- y x))
(defunalias eq! eq)
(defconstraint test () (eq! A B))
