(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns ST (A :i16@loob) B C)
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! B)
      (vanishes! C)))
