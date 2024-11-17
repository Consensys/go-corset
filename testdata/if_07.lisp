(defcolumns X Y Z)
(defconstraint test () (if X 0 (- Z (if Y 3 16))))
