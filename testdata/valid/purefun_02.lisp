(defcolumns (A :i16) (B :i16))
(defpurefun ((eq :bool) x y) (== y x))
(defconstraint test () (eq A B))
