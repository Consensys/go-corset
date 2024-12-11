(defpurefun ((vanishes! :@loob) x) x)

(defcolumns ST (A :@loob) B)
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! 0)
      (vanishes! B)))
