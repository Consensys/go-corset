(defpurefun ((vanishes! :@loob) x) x)

(defcolumns ST (A :@loob) B C)
(defconstraint c1 (:guard ST)
  (if A
      (vanishes! B)
      (vanishes! C)))
