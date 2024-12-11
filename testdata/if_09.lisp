(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test () (- Z (if X (if Y 0 16))))
