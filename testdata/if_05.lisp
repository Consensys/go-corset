(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test () (if X (- Z (if Y 0 16))))
