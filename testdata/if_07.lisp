(defcolumns (X :@loob) (Y :@loob) Z)
(defconstraint test () (if X 0 (- Z (if Y 3 16))))
