(defcolumns X Y Z)
(defconstraint test () (ifnot X (- Z (if Y 3 16))))
