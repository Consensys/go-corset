(defcolumns X Y Z)
(defconstraint test () (if X (- Z (if Y 0))))
(defconstraint test () (if X (- Z (ifnot Y 16))))
