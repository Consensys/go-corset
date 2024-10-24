(defcolumns ST X Y Z)
(defconstraint test (:guard ST) (if X (- Z (if Y 0 16))))
