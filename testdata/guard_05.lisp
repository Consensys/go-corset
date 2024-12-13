(defpurefun ((vanishes! :@loob) x) x)

(defcolumns ST (X :@loob) (Y :@loob) (Z :@loob))
(defconstraint test (:guard ST)
  (if X
      (vanishes! (- Z (if Y 0 16)))))
