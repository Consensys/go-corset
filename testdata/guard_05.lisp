(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns ST (X :i16@loob) (Y :i16@loob) (Z :i16@loob))
(defconstraint test (:guard ST)
  (if X
      (vanishes! (- Z (if Y 0 16)))))
