(defcolumns (A :i16) (B :i16))
(defpurefun ((eq :i16@loob) x y) (- y x))
(defconstraint test () (eq A B))
