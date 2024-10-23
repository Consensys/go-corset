(defcolumns X Y Z)
(defconstraint test () (- Z (if X (if Y 0 16))))
