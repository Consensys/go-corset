(defpurefun ((vanishes! :𝔽@loob) x) x)

(defcolumns ST (A :i16@loob) B)
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! 0)
      (vanishes! B)))
