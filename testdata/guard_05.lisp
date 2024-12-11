(defcolumns ST (X :@loob) (Y :@loob) (Z :@loob))
(defconstraint test (:guard ST) (if X (- Z (if Y 0 16))))
